package utils

import (
	"os"
	"time"
)

// FixedTime используется для тестов, если не nil.
var FixedTime *time.Time

// Now возвращает текущее время UTC, обрезанное до начала дня.
// Если установлена переменная окружения FIXED_TIME (формат "20060102"), то используется её значение.
func Now() time.Time {
	if FixedTime != nil {
		return *FixedTime
	}
	if ft := os.Getenv("FIXED_TIME"); ft != "" {
		if t, err := time.Parse("20060102", ft); err == nil {
			FixedTime = &t
			return t
		}
	}
	return time.Now().UTC().Truncate(24 * time.Hour)
}
