package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	Pool         *pgxpool.Pool
	RedisClient  *redis.Client
	ReadyTimeout time.Duration
}

func NewHealthHandler(pool *pgxpool.Pool,
	redisClient *redis.Client,
	readyTimeout time.Duration) *HealthHandler {
	return &HealthHandler{
		RedisClient:  redisClient,
		Pool:         pool,
		ReadyTimeout: readyTimeout,
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(map[string]string{"status": "ok"})
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return
	}
}
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	readyCtx, cancel := context.WithTimeout(r.Context(), h.ReadyTimeout)
	defer cancel()
	err := h.Pool.Ping(readyCtx)
	if err != nil {
		status := http.StatusServiceUnavailable
		http.Error(w, http.StatusText(status), status)
		return
	}
	err = h.RedisClient.Ping(readyCtx).Err()
	if err != nil {
		status := http.StatusServiceUnavailable
		http.Error(w, http.StatusText(status), status)
		return
	}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(map[string]string{"status": "ready"})
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return
	}
}
