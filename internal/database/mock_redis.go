package database

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	mockRedis     *MockRedis
	mockRedisOnce sync.Once
	UseMockRedis  bool
)

type MockRedis struct {
	mu   sync.RWMutex
	data map[string]string
	ttls map[string]time.Time
}

func initMockRedis() {
	mockRedisOnce.Do(func() {
		mockRedis = &MockRedis{
			data: make(map[string]string),
			ttls: make(map[string]time.Time),
		}
	})
}

func GetMockRedis() *MockRedis {
	initMockRedis()
	return mockRedis
}

func (m *MockRedis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[key]; exists {
		if expiry, hasTTL := m.ttls[key]; hasTTL {
			if time.Now().After(expiry) {
				delete(m.data, key)
				delete(m.ttls, key)
			} else {
				return redis.NewBoolResult(false, nil)
			}
		} else {
			return redis.NewBoolResult(false, nil)
		}
	}

	m.data[key] = value.(string)
	if expiration > 0 {
		m.ttls[key] = time.Now().Add(expiration)
	}

	return redis.NewBoolResult(true, nil)
}

func (m *MockRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	deleted := 0
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			delete(m.data, key)
			delete(m.ttls, key)
			deleted++
		}
	}

	return redis.NewIntResult(int64(deleted), nil)
}

func (m *MockRedis) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, key := range keys {
		if val, exists := m.data[key]; exists {
			if expiry, hasTTL := m.ttls[key]; hasTTL && time.Now().After(expiry) {
				continue
			}
			if val != "" {
				count++
			}
		}
	}

	return redis.NewIntResult(int64(count), nil)
}

func (m *MockRedis) Ping(ctx context.Context) *redis.StatusCmd {
	return redis.NewStatusResult("PONG", nil)
}

func (m *MockRedis) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]string)
	m.ttls = make(map[string]time.Time)
}
