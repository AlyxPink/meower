package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lua script for atomic rate limiting
// Returns: {allowed (0/1), remaining_tokens}
const rateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local window_start = now - window

-- Remove old entries outside the time window
redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

-- Get current count of requests in window
local count = redis.call('ZCARD', key)

-- Check if limit would be exceeded
if count >= limit then
    return {0, 0}  -- Not allowed, no remaining tokens
end

-- Add current request to the sorted set
redis.call('ZADD', key, now, tostring(now))

-- Set expiry on the key (window + buffer for cleanup)
redis.call('EXPIRE', key, window)

-- Calculate remaining tokens
local remaining = limit - count - 1
return {1, remaining}  -- Allowed, with remaining count
`

// RateLimiter implements token bucket rate limiting using Redis
type RateLimiter struct {
	redis *redis.Client
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisURL string) (*RateLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RateLimiter{
		redis: client,
	}, nil
}

// Allow checks if a request is allowed based on rate limit
// Uses sliding window algorithm with atomic Lua script execution
// Returns: allowed (bool), remaining tokens (int), error
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	now := time.Now().Unix()
	windowSeconds := int64(window.Seconds())

	// Execute Lua script atomically
	result, err := rl.redis.Eval(ctx, rateLimitScript, []string{key}, limit, windowSeconds, now).Result()
	if err != nil {
		return false, 0, fmt.Errorf("rate limit check failed: %w", err)
	}

	// Parse result from Lua script
	resultSlice, ok := result.([]any)
	if !ok || len(resultSlice) != 2 {
		return false, 0, fmt.Errorf("unexpected result format from rate limit script")
	}

	// Extract allowed flag and remaining count
	allowed, ok := resultSlice[0].(int64)
	if !ok {
		return false, 0, fmt.Errorf("invalid allowed flag in rate limit result")
	}

	remaining, ok := resultSlice[1].(int64)
	if !ok {
		return false, 0, fmt.Errorf("invalid remaining count in rate limit result")
	}

	return allowed == 1, int(remaining), nil
}

// Close closes the Redis connection
func (rl *RateLimiter) Close() error {
	return rl.redis.Close()
}
