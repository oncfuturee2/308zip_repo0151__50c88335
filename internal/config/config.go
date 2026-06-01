package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	RedisHost             string
	RedisPort             string
	RedisPassword         string
	JWTSecret             string
	JWTExpireHours        int
	CommissionRateLevel1  int
	CommissionRateLevel2  int
	CommissionRateLevel3  int
}

var AppConfig *Config

func LoadConfig() {
	_ = godotenv.Load()

	AppConfig = &Config{
		DBHost:                getEnv("DB_HOST", "localhost"),
		DBPort:                getEnv("DB_PORT", "5432"),
		DBUser:                getEnv("DB_USER", "postgres"),
		DBPassword:            getEnv("DB_PASSWORD", "postgres123"),
		DBName:                getEnv("DB_NAME", "distribution"),
		RedisHost:             getEnv("REDIS_HOST", "localhost"),
		RedisPort:             getEnv("REDIS_PORT", "6379"),
		RedisPassword:         getEnv("REDIS_PASSWORD", ""),
		JWTSecret:             getEnv("JWT_SECRET", "your_jwt_secret_key_here"),
		JWTExpireHours:        getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		CommissionRateLevel1:  getEnvAsInt("COMMISSION_RATE_LEVEL1", 10),
		CommissionRateLevel2:  getEnvAsInt("COMMISSION_RATE_LEVEL2", 5),
		CommissionRateLevel3:  getEnvAsInt("COMMISSION_RATE_LEVEL3", 3),
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
