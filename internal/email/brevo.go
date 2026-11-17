package email

import (
	"bytes"
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var _ Sender = (*BrevoSender)(nil)

const brevoSendEmailURL = "https://api.brevo.com/v3/smtp/email"

var (
	ErrBrevoAPIKeyEmpty      = errors.New("Brevo API key is empty")
	ErrBrevoSenderEmailEmpty = errors.New("Brevo sender email is empty")
	ErrBrevoHTTPClientNil    = errors.New("Brevo HTTP client is nil")
)

type verificationPayload struct {
	Code     string `json:"code"`
	Username string `json:"username"`
}

type brevoAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type brevoRequest struct {
	Sender      brevoAddress      `json:"sender"`
	To          []brevoAddress    `json:"to"`
	Subject     string            `json:"subject"`
	TextContent string            `json:"textContent"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type BrevoSender struct {
	apiKey      string
	senderEmail string
	senderName  string
	endpoint    string
	client      *http.Client
}

func NewBrevoSender(apiKey string, senderEmail string, senderName string, client *http.Client) (*BrevoSender, error) {
	apiKey = strings.TrimSpace(apiKey)
	senderEmail = strings.TrimSpace(senderEmail)
	senderName = strings.TrimSpace(senderName)

	if apiKey == "" {
		return nil, ErrBrevoAPIKeyEmpty
	}
	if senderEmail == "" {
		return nil, ErrBrevoSenderEmailEmpty
	}
	if client == nil {
		return nil, ErrBrevoHTTPClientNil
	}
	if senderName == "" {
		senderName = "Do Together"
	}

	return &BrevoSender{
		apiKey:      apiKey,
		senderEmail: senderEmail,
		senderName:  senderName,
		endpoint:    brevoSendEmailURL,
		client:      client,
	}, nil
}

func (s *BrevoSender) Send(ctx context.Context, message *domain.EmailOutboxMessage) error {
	if message == nil {
		return repository.ErrEmailOutboxMessageNil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if message.Kind != domain.EmailKindVerification {
		return fmt.Errorf("unsupported email kind: %s", message.Kind)
	}

	var payload verificationPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return fmt.Errorf("decode verification email payload: %w", err)
	}
	if strings.TrimSpace(payload.Code) == "" {
		return errors.New("verification email code is empty")
	}

	body := brevoRequest{
		Sender: brevoAddress{
			Email: s.senderEmail,
			Name:  s.senderName,
		},
		To: []brevoAddress{
			{
				Email: message.Recipient,
				Name:  payload.Username,
			},
		},
		Subject: "Подтверждение email в Do Together",
		TextContent: fmt.Sprintf(
			"Здравствуйте, %s!\n\nКод подтверждения: %s\n\nЕсли вы не регистрировались в Do Together, проигнорируйте это письмо.",
			payload.Username,
			payload.Code,
		),
		Headers: map[string]string{
			"idempotencyKey": message.ID.String(),
		},
	}
	requestBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode Brevo request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.endpoint,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return fmt.Errorf("create Brevo request: %w", err)
	}

	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")

	response, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send Brevo request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf(
			"Brevo returned status %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}

	return nil
}
