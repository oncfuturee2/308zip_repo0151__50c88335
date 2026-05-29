package database

import (
	"context"
	"fmt"
	"log"

	"distribution-commission/internal/config"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	addr := fmt.Sprintf("%s:%s", config.AppConfig.RedisHost, config.AppConfig.RedisPort)
	
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.AppConfig.RedisPassword,
		DB:       0,
	})

	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
}
