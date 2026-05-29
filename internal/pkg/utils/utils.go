package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/now"
)

func GenerateOrderNo() string {
	prefix := "ORD"
	timestamp := now.BeginningOfDay().Format("20060102")
	random := uuid.New().String()[:8]
	return fmt.Sprintf("%s%s%s", prefix, timestamp, random)
}

func GenerateWithdrawRequestNo() string {
	prefix := "WTH"
	timestamp := time.Now().Format("20060102150405")
	random := uuid.New().String()[:6]
	return fmt.Sprintf("%s%s%s", prefix, timestamp, random)
}

func FormatAmount(amount int64) string {
	yuan := float64(amount) / 100.0
	return fmt.Sprintf("%.2f", yuan)
}
