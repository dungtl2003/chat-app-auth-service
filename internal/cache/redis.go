package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	internal *redis.UniversalClient
}

type Config struct {
	Addrs    []string
	Password string
}

func NewRedisClient(ctx context.Context, cfg Config) (*Client, error) {
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    cfg.Addrs,
		Password: cfg.Password,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{
		internal: &rdb,
	}, nil
}

func (c *Client) GetPasswordResetRateLimitKey(email string) string {
	return "pwd_reset_rl:" + email
}

func (c *Client) IncrPasswordResetRateLimit(ctx context.Context, email string) (int64, error) {
	key := c.GetPasswordResetRateLimitKey(email)
	val, err := (*c.internal).Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Client) DecrPasswordResetRateLimit(ctx context.Context, email string) (int64, error) {
	key := c.GetPasswordResetRateLimitKey(email)
	val, err := (*c.internal).Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Client) ExpirePasswordResetRateLimit(ctx context.Context, email string, ttl time.Duration) error {
	key := c.GetPasswordResetRateLimitKey(email)
	return (*c.internal).Expire(ctx, key, ttl).Err()
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
	return (*c.internal).Set(ctx, key, code, ttl).Err()
}

func (c *Client) GetPasswordResetCode(ctx context.Context, email string) (string, error) {
	key := c.GetPasswordResetCodeKey(email)
	val, err := (*c.internal).Get(ctx, key).Result()
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
	return (*c.internal).Del(ctx, key).Err()
}

func (c *Client) GetPasswordResetAttemptKey(email string) string {
	return "pwd_reset_attempts:" + email
}

func (c *Client) IncrPasswordResetAttempt(ctx context.Context, email string) (int64, error) {
	key := c.GetPasswordResetAttemptKey(email)
	val, err := (*c.internal).Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Client) DecrPasswordResetAttempt(ctx context.Context, email string) (int64, error) {
	key := c.GetPasswordResetAttemptKey(email)
	val, err := (*c.internal).Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Client) ExpirePasswordResetAttempt(ctx context.Context, email string, ttl time.Duration) error {
	key := c.GetPasswordResetAttemptKey(email)
	return (*c.internal).Expire(ctx, key, ttl).Err()
}

func (c *Client) FlushAll(ctx context.Context) error {
	return (*c.internal).FlushAll(ctx).Err()
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
	return (*c.internal).Set(ctx, key, token, ttl).Err()
}

func (c *Client) GetPasswordResetToken(ctx context.Context, email string) (string, error) {
	key := c.GetPasswordResetTokenKey(email)
	val, err := (*c.internal).Get(ctx, key).Result()
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
	return (*c.internal).Del(ctx, key).Err()
}
