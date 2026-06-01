package database

import (
	"log"

	"distribution-commission/internal/models"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testRedisServer *miniredis.Miniredis

func SetupTestDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Distributor{},
		&models.Order{},
		&models.CommissionRecord{},
		&models.WithdrawRequest{},
		&models.AuditLog{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	SetupTestRedis()
}

func CleanupTestDB() {
	DB.Exec("DELETE FROM audit_logs")
	DB.Exec("DELETE FROM withdraw_requests")
	DB.Exec("DELETE FROM commission_records")
	DB.Exec("DELETE FROM orders")
	DB.Exec("DELETE FROM distributors")
	DB.Exec("DELETE FROM users")

	if RedisClient != nil {
		_ = RedisClient.Close()
		RedisClient = nil
	}

	if testRedisServer != nil {
		testRedisServer.Close()
		testRedisServer = nil
	}
}

func SetupTestRedis() {
	if testRedisServer != nil {
		testRedisServer.Close()
	}

	server, err := miniredis.Run()
	if err != nil {
		log.Fatalf("Failed to start test redis: %v", err)
	}

	testRedisServer = server
	RedisClient = redis.NewClient(&redis.Options{
		Addr: server.Addr(),
		DB:   0,
	})
}
