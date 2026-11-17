package main

import (
	"context"
	"do-together/internal/email"
	"do-together/internal/repository/postgres"
	"do-together/internal/service"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	pollInterval  = 5 * time.Second
	batchSize     = 10
	leaseDuration = 20 * time.Second
	httpTimeout   = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	databaseURL, err := requiredEnv("DATABASE_URL")
	if err != nil {
		return err
	}
	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	var sender email.Sender
	transport := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_TRANSPORT")))
	if transport == "" || transport == "log" {
		sender = email.NewLogSender()
	} else if transport == "brevo" {
		brevoAPIKey, err := requiredEnv("BREVO_API_KEY")
		if err != nil {
			return err
		}
		brevoSenderEmail, err := requiredEnv("BREVO_SENDER_EMAIL")
		if err != nil {
			return err
		}
		sender, err = email.NewBrevoSender(brevoAPIKey, brevoSenderEmail, strings.TrimSpace(os.Getenv("BREVO_SENDER_NAME")), &http.Client{Timeout: httpTimeout})
		if err != nil {
			return fmt.Errorf("create Brevo sender: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported EMAIL_TRANSPORT %q", transport)
	}

	workerID, err := newWorkerID()
	if err != nil {
		return err
	}
	outboxRepository := postgres.NewPostgresEmailOutboxRepository(pool)
	worker, err := service.NewEmailWorker(
		outboxRepository,
		sender,
		workerID,
		batchSize,
		leaseDuration,
	)
	if err != nil {
		return fmt.Errorf("create email worker: %w", err)
	}

	log.Printf("email worker started: %s", workerID)
	processBatch(ctx, worker)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("email worker shutdown signal received")
			return nil
		case <-ticker.C:
			processBatch(ctx, worker)
		}
	}
}

func processBatch(ctx context.Context, worker *service.EmailWorker) {
	processed, err := worker.ProcessBatch(ctx)
	if err != nil {
		log.Printf("process email batch: %v", err)
	}
	if processed > 0 {
		log.Printf("processed email messages: %d", processed)
	}
}

func requiredEnv(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func newWorkerID() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("get hostname: %w", err)
	}
	return fmt.Sprintf("%s-%d", hostname, os.Getpid()), nil
}
