package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	RedisHost            string
	RedisPort            string
	RedisPassword        string
	JWTSecret            string
	JWTExpireHours       int
	CommissionRate       int
	Level1CommissionRate int
	Level2CommissionRate int
	Level3CommissionRate int
}

var AppConfig *Config

func LoadConfig() {
	_ = godotenv.Load()

	AppConfig = &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgres123"),
		DBName:               getEnv("DB_NAME", "distribution"),
		RedisHost:            getEnv("REDIS_HOST", "localhost"),
		RedisPort:            getEnv("REDIS_PORT", "6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		JWTSecret:            getEnv("JWT_SECRET", "your_jwt_secret_key_here"),
		JWTExpireHours:       getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		CommissionRate:       getEnvAsInt("COMMISSION_RATE", 10),
		Level1CommissionRate: getEnvAsInt("LEVEL1_COMMISSION_RATE", 10),
		Level2CommissionRate: getEnvAsInt("LEVEL2_COMMISSION_RATE", 5),
		Level3CommissionRate: getEnvAsInt("LEVEL3_COMMISSION_RATE", 2),
	}

	log.Println("Config loaded successfully")
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}
