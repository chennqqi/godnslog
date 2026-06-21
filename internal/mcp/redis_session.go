package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisSessionStore implements SessionStore-like functionality using Redis as the backend.
// This enables session sharing across multiple MCP server instances in HA deployments.
type RedisSessionStore struct {
	client    *redis.Client
	timeout   time.Duration
	keyPrefix string
}

// NewRedisSessionStore creates a new Redis-backed session store.
// If client is nil, returns nil (caller should fall back to in-memory store).
func NewRedisSessionStore(client *redis.Client, timeout time.Duration) *RedisSessionStore {
	if client == nil {
		return nil
	}
	return &RedisSessionStore{
		client:    client,
		timeout:   timeout,
		keyPrefix: "mcp:session:",
	}
}

// redisSessionData is the serialized form of a session stored in Redis.
type redisSessionData struct {
	ID        string                 `json:"id"`
	CreatedAt time.Time              `json:"created_at"`
	LastSeen  time.Time              `json:"last_seen"`
	Data      map[string]interface{} `json:"data"`
}

// Create creates a new session in Redis and returns it.
func (s *RedisSessionStore) Create() *Session {
	session := &Session{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		data:      make(map[string]interface{}),
	}

	// Persist to Redis
	data := redisSessionData{
		ID:        session.ID,
		CreatedAt: session.CreatedAt,
		LastSeen:  session.LastSeen,
		Data:      make(map[string]interface{}),
	}
	jsonData, _ := json.Marshal(data)
	ctx := context.Background()
	s.client.Set(ctx, s.keyPrefix+session.ID, jsonData, s.timeout)

	return session
}

// Get retrieves a session by ID from Redis. Returns nil if not found or expired.
func (s *RedisSessionStore) Get(id string) *Session {
	ctx := context.Background()
	jsonData, err := s.client.Get(ctx, s.keyPrefix+id).Bytes()
	if err != nil {
		return nil
	}

	var data redisSessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil
	}

	// Check if session has expired
	if time.Since(data.LastSeen) > s.timeout {
		s.Delete(id)
		return nil
	}

	session := &Session{
		ID:        data.ID,
		CreatedAt: data.CreatedAt,
		LastSeen:  data.LastSeen,
		data:      data.Data,
	}
	if session.data == nil {
		session.data = make(map[string]interface{})
	}

	// Update last seen time
	session.Touch()
	s.persist(session)

	return session
}

// Delete removes a session from Redis.
func (s *RedisSessionStore) Delete(id string) {
	ctx := context.Background()
	s.client.Del(ctx, s.keyPrefix+id)
}

// Cleanup is a no-op for Redis store since Redis handles TTL automatically.
func (s *RedisSessionStore) Cleanup() {
	// Redis TTL handles expiration automatically
}

// StartCleanupRoutine is a no-op for Redis store.
func (s *RedisSessionStore) StartCleanupRoutine(ctx context.Context) {
	// Redis TTL handles expiration automatically
}

// persist saves the session state to Redis with updated TTL.
func (s *RedisSessionStore) persist(session *Session) {
	data := redisSessionData{
		ID:        session.ID,
		CreatedAt: session.CreatedAt,
		LastSeen:  session.LastSeen,
		Data:      session.data,
	}
	jsonData, _ := json.Marshal(data)
	ctx := context.Background()
	s.client.Set(ctx, s.keyPrefix+session.ID, jsonData, s.timeout)
}

// Set stores a value in the session and persists to Redis.
func (s *RedisSessionStore) Set(session *Session, key string, value interface{}) {
	session.Set(key, value)
	s.persist(session)
}

// sessionStoreInterface defines the interface that both in-memory and Redis session stores implement.
type sessionStoreInterface interface {
	Create() *Session
	Get(id string) *Session
	Delete(id string)
	Cleanup()
	StartCleanupRoutine(ctx context.Context)
}

// Ensure RedisSessionStore implements sessionStoreInterface
var _ sessionStoreInterface = (*RedisSessionStore)(nil)

// Ensure SessionStore implements sessionStoreInterface
var _ sessionStoreInterface = (*SessionStore)(nil)

// newSessionStore creates a Redis-backed store if client is provided, otherwise falls back to in-memory.
func newSessionStore(timeout time.Duration, redisClient *redis.Client) sessionStoreInterface {
	if redisClient != nil {
		return NewRedisSessionStore(redisClient, timeout)
	}
	return NewSessionStore(timeout)
}

// formatRedisKey returns the Redis key for a session ID.
func (s *RedisSessionStore) formatRedisKey(id string) string {
	return fmt.Sprintf("%s%s", s.keyPrefix, id)
}
