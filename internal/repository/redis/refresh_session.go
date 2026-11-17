package redis

import (
	"context"
	"do-together/internal/domain"
	"do-together/internal/repository"
	"errors"
	"fmt"
	"time"
	"uuid"

	goredis "github.com/redis/go-redis/v9"
)

// var _ repository.RefreshSessionRepository = (*RefreshSessionRepository)(nil)
var (
	ErrSessionNil = errors.New("session is nil")
)
var createRefreshSessionScript = goredis.NewScript(`
	if redis.call("EXISTS", KEYS[1]) == 1 then
		return 0
	end

	redis.call(
		"HSET",
		KEYS[1],
		"user_id", ARGV[1],
		"token_hash", ARGV[2],
		"created_at", ARGV[3],
		"absolute_expires_at", ARGV[4]
	)

	redis.call("PEXPIRE", KEYS[1], ARGV[5])

	return 1
`)
var rotateRefreshSessionScript = goredis.NewScript(`
	local values = redis.call(
		"HMGET",
		KEYS[1],
		"user_id",
		"token_hash",
		"absolute_expires_at"
	)

	if not values[1] or not values[2] or not values[3] then
		return -1
	end

	local userID = values[1]
	local storedHash = values[2]
	local absoluteExpiresAt = tonumber(values[3])

	if storedHash ~= ARGV[1] then
		return -2
	end

	local now = tonumber(ARGV[3])
	local idleTTL = tonumber(ARGV[4])
	local remaining = absoluteExpiresAt - now

	if remaining <= 0 then
		redis.call("DEL", KEYS[1])
		return -3
	end

	local newTTL = math.min(idleTTL, remaining)

	redis.call("HSET", KEYS[1], "token_hash", ARGV[2])
	redis.call("PEXPIRE", KEYS[1], newTTL)

	return {4, userID}
`)

type RefreshSessionRepository struct {
	client *goredis.Client
}

const refreshSessionKeyPrefix = "refresh_session:"

func NewRefreshSessionRepository(client *goredis.Client) *RefreshSessionRepository {
	return &RefreshSessionRepository{
		client: client,
	}
}

func redisKeyFromSessionID(sessionID string) string {
	return refreshSessionKeyPrefix + sessionID
}

func (r *RefreshSessionRepository) Create(ctx context.Context, session *domain.RefreshSession, idleTTL time.Duration) error {
	if session == nil {
		return ErrSessionNil
	}
	key := redisKeyFromSessionID(session.SessionID)
	if err := ctx.Err(); err != nil {
		return err
	}
	if idleTTL <= 0 {
		return fmt.Errorf("idle TTL must be positive")
	}
	result, err := createRefreshSessionScript.Run(ctx, r.client, []string{key}, session.UserID.String(), session.TokenHash, session.CreatedAt.UnixMilli(), session.AbsoluteExpiresAt.UnixMilli(), idleTTL.Milliseconds()).Int()
	if err != nil {
		return errors.Join(
			repository.ErrRedisUnavailable,
			fmt.Errorf("create refresh session: %w", err),
		)
	}

	if result == 0 {
		return repository.ErrSessionAlreadyUsed
	}
	if result != 1 {
		return fmt.Errorf("unexpected create refresh session result: %d", result)
	}

	return nil
}
func (r *RefreshSessionRepository) Rotate(ctx context.Context, sessionID, oldHash, newHash string, now time.Time, idleTTL time.Duration) (userID uuid.UUID, err error) {
	if err := ctx.Err(); err != nil {
		return uuid.Nil(), err
	}
	if sessionID == "" || oldHash == "" || newHash == "" {
		return uuid.Nil(), fmt.Errorf("refresh session rotation arguments are required")
	}
	if idleTTL <= 0 {
		return uuid.Nil(), fmt.Errorf("idle TTL must be positive")
	}
	key := redisKeyFromSessionID(sessionID)
	result, err := rotateRefreshSessionScript.Run(
		ctx,
		r.client,
		[]string{key},
		oldHash,
		newHash,
		now.UnixMilli(),
		idleTTL.Milliseconds(),
	).Slice()
	if err != nil {
		return uuid.Nil(), errors.Join(
			repository.ErrRedisUnavailable,
			fmt.Errorf("rotate refresh session: %w", err),
		)
	}
	if len(result) == 1 {
		code, _ := result[0].(int64)
		switch code {
		case -1:
			return uuid.Nil(), repository.ErrSessionNotFound
		case -2:
			return uuid.Nil(), repository.ErrTokenMismatch
		case -3:
			return uuid.Nil(), repository.ErrSessionExpired
		}
	}
	if len(result) != 2 {
		return uuid.Nil(), fmt.Errorf("unexpected rotate refresh session result: %v", result)
	}
	value, ok := result[1].(string)
	if !ok {
		return uuid.Nil(), fmt.Errorf("unexpected refresh user id: %v", result[1])
	}
	parsed, parseErr := uuid.Parse(value)
	if parseErr != nil || parsed == uuid.Nil() {
		return uuid.Nil(), fmt.Errorf("invalid refresh user id")
	}
	return parsed, nil
}
func (r *RefreshSessionRepository) Delete(ctx context.Context, sessionID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	key := redisKeyFromSessionID(sessionID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return errors.Join(
			repository.ErrRedisUnavailable,
			fmt.Errorf("delete refresh session: %w", err),
		)
	}

	return nil
}
