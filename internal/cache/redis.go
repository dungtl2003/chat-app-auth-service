package cache

import (
	"context"
	"dungtl2003/chat-app-auth-service/internal/logging"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	writer redis.UniversalClient
	reader redis.UniversalClient
	logger *logging.LoggerWrapper
}

type Config struct {
	MasterAddr  string
	ReplicaAddr string
	Password    string
	Logger      *logging.LoggerWrapper
}

func NewRedisClient(ctx context.Context, cfg Config) (*Client, error) {
	writer := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{cfg.MasterAddr},
		Password: cfg.Password,
	})

	reader := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{cfg.ReplicaAddr},
		Password: cfg.Password,
	})

	if err := writer.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	if err := reader.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{
		writer: writer,
		reader: reader,
		logger: cfg.Logger,
	}, nil

}

func (c *Client) GetPasswordResetRateLimitKey(email string) string {
	return "pwd_reset_rl:" + email
}

func (c *Client) IncrPasswordResetRateLimit(ctx context.Context, email string) (int64, error) {
	key := c.GetPasswordResetRateLimitKey(email)
	val, err := c.writer.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Client) DecrPasswordResetRateLimit(ctx context.Context, email string) (int64, error) {
	key := c.GetPasswordResetRateLimitKey(email)
	val, err := c.writer.Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Client) ExpirePasswordResetRateLimit(ctx context.Context, email string, ttl time.Duration) error {
	key := c.GetPasswordResetRateLimitKey(email)
	return c.writer.Expire(ctx, key, ttl).Err()
}

func (c *Client) GetPasswordResetCodeKey(email string) string {
	return "pwd_reset_code:" + email
}

func (c *Client) SetPasswordResetCode(
	ctx context.Context,
	email, code string,
	ttl time.Duration,
) error {
	key := c.GetPasswordResetCodeKey(email)
	return c.writer.Set(ctx, key, code, ttl).Err()
}

func (c *Client) GetPasswordResetCode(ctx context.Context, email string) (string, error) {
	key := c.GetPasswordResetCodeKey(email)
	val, err := c.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *Client) DeletePasswordResetCode(ctx context.Context, email string) error {
	key := c.GetPasswordResetCodeKey(email)
	return c.writer.Del(ctx, key).Err()
}

func (c *Client) GetPasswordResetAttemptKey(email string) string {
	return "pwd_reset_attempts:" + email
}

func (c *Client) IncrPasswordResetAttempt(ctx context.Context, email string) (int64, error) {
	key := c.GetPasswordResetAttemptKey(email)
	val, err := c.writer.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Client) DecrPasswordResetAttempt(ctx context.Context, email string) (int64, error) {
	key := c.GetPasswordResetAttemptKey(email)
	val, err := c.writer.Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Client) DeletePasswordResetAttempt(ctx context.Context, email string) error {
	key := c.GetPasswordResetAttemptKey(email)
	return c.writer.Del(ctx, key).Err()
}

func (c *Client) ExpirePasswordResetAttempt(ctx context.Context, email string, ttl time.Duration) error {
	key := c.GetPasswordResetAttemptKey(email)
	return c.writer.Expire(ctx, key, ttl).Err()
}

func (c *Client) FlushAll(ctx context.Context) error {
	return c.writer.FlushAll(ctx).Err()
}

func (c *Client) GetPasswordResetTokenKey(email string) string {
	return "pwd_reset_token:" + email
}

func (c *Client) SetPasswordResetToken(
	ctx context.Context,
	email, token string,
	ttl time.Duration,
) error {
	key := c.GetPasswordResetTokenKey(email)
	return c.writer.Set(ctx, key, token, ttl).Err()
}

func (c *Client) GetPasswordResetToken(ctx context.Context, email string) (string, error) {
	key := c.GetPasswordResetTokenKey(email)
	val, err := c.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *Client) DeletePasswordResetToken(ctx context.Context, email string) error {
	key := c.GetPasswordResetTokenKey(email)
	return c.writer.Del(ctx, key).Err()
}

func (c *Client) GetBlacklistSessionKey(sessionId int64) string {
	return fmt.Sprintf("blacklist_session_%d", sessionId)
}

func (c *Client) IsSessionBlacklisted(ctx context.Context, sessionId int64) (bool, error) {
	key := c.GetBlacklistSessionKey(sessionId)
	val, err := c.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

func (c *Client) Get(ctx context.Context, key string) *redis.StringCmd {
	val, err := c.reader.Get(ctx, key).Result()
	if err != nil && isNetworkError(err) {
		c.logger.Warn("Replicas down! Falling back to Master for read operation.")
		return c.writer.Get(ctx, key)
	}

	return redis.NewStringResult(val, err)
}

func isNetworkError(err error) bool {
	return err != redis.Nil // redis.Nil means the key doesn't exist, which is a normal response
}
