package service

import (
	"context"
	"fmt"
	"time"

	"distribution-commission/internal/database"
)

const (
	IdempotentKeyPrefixOrder      = "idempotent:order"
	IdempotentKeyPrefixCommission = "idempotent:commission"
	IdempotentKeyTTL              = 24 * time.Hour
)

type IdempotentService struct{}

func NewIdempotentService() *IdempotentService {
	return &IdempotentService{}
}

func (s *IdempotentService) buildKey(prefix, identifier string) string {
	return fmt.Sprintf("%s:%s", prefix, identifier)
}

func (s *IdempotentService) AcquireOrderLock(ctx context.Context, orderNo string) (bool, error) {
	key := s.buildKey(IdempotentKeyPrefixOrder, orderNo)
	return s.acquireLock(ctx, key)
}

func (s *IdempotentService) ReleaseOrderLock(ctx context.Context, orderNo string) error {
	key := s.buildKey(IdempotentKeyPrefixOrder, orderNo)
	return s.releaseLock(ctx, key)
}

func (s *IdempotentService) AcquireCommissionLock(ctx context.Context, orderNo string) (bool, error) {
	key := s.buildKey(IdempotentKeyPrefixCommission, orderNo)
	return s.acquireLock(ctx, key)
}

func (s *IdempotentService) ReleaseCommissionLock(ctx context.Context, orderNo string) error {
	key := s.buildKey(IdempotentKeyPrefixCommission, orderNo)
	return s.releaseLock(ctx, key)
}

func (s *IdempotentService) acquireLock(ctx context.Context, key string) (bool, error) {
	var result bool
	var err error

	if database.UseMockRedis {
		result, err = database.GetMockRedis().SetNX(ctx, key, "1", IdempotentKeyTTL).Result()
	} else {
		result, err = database.RedisClient.SetNX(ctx, key, "1", IdempotentKeyTTL).Result()
	}

	if err != nil {
		return false, err
	}
	return result, nil
}

func (s *IdempotentService) releaseLock(ctx context.Context, key string) error {
	if database.UseMockRedis {
		return database.GetMockRedis().Del(ctx, key).Err()
	}
	return database.RedisClient.Del(ctx, key).Err()
}

func (s *IdempotentService) IsOrderProcessed(ctx context.Context, orderNo string) (bool, error) {
	key := s.buildKey(IdempotentKeyPrefixOrder, orderNo)
	exists, err := s.checkExists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *IdempotentService) IsCommissionGenerated(ctx context.Context, orderNo string) (bool, error) {
	key := s.buildKey(IdempotentKeyPrefixCommission, orderNo)
	exists, err := s.checkExists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *IdempotentService) checkExists(ctx context.Context, key string) (int64, error) {
	if database.UseMockRedis {
		return database.GetMockRedis().Exists(ctx, key).Result()
	}
	return database.RedisClient.Exists(ctx, key).Result()
}
