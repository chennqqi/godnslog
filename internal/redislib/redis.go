// Package redislib provides a common Redis client wrapper for shared use
// across MCP session store, workflow queue, and HA modules.
package redislib

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis connection configuration.
type Config struct {
	Addr     string
	Password string
	DB       int
	UseTLS   bool
}

// Client wraps a go-redis client with key prefix management.
type Client struct {
	raw    *redis.Client
	prefix string
}

// NewClient creates a new Redis client. Returns nil if addr is empty.
func NewClient(cfg Config) *Client {
	if cfg.Addr == "" {
		return nil
	}
	opts := &redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	return &Client{
		raw:    redis.NewClient(opts),
		prefix: "godnslog:",
	}
}

// Ping checks the Redis connection health.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.raw == nil {
		return fmt.Errorf("redis client is nil")
	}
	return c.raw.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	if c == nil || c.raw == nil {
		return nil
	}
	return c.raw.Close()
}

// Raw returns the underlying go-redis client for direct use.
func (c *Client) Raw() *redis.Client {
	if c == nil {
		return nil
	}
	return c.raw
}

// Key returns a prefixed Redis key.
func (c *Client) Key(suffix string) string {
	return c.prefix + suffix
}
