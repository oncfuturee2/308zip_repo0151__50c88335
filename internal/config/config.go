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
	CommissionLevel2Rate int
	CommissionLevel3Rate int
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
		CommissionLevel2Rate: getEnvAsInt("COMMISSION_LEVEL2_RATE", 5),
		CommissionLevel3Rate: getEnvAsInt("COMMISSION_LEVEL3_RATE", 2),
	}

	log.Println("Config loaded successfully")
}

func (c *Config) CommissionRates() []int {
	if c == nil {
		return []int{10, 5, 2}
	}

	return []int{c.CommissionRate, c.CommissionLevel2Rate, c.CommissionLevel3Rate}
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
