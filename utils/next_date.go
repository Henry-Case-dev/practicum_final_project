package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	log.Printf("NextDate called with now: %v, dateStr: %s, repeat: %s", now, dateStr, repeat)
	if repeat == "" {
		return "", nil
	}

	parsedDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date: %v", err)
	}
	log.Printf("Parsed date: %v", parsedDate)

	switch {
	case strings.HasPrefix(repeat, "d "):
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid repeat format")
		}

		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("invalid days value")
		}

		nextDate := parsedDate.AddDate(0, 0, days)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}
		log.Printf("Next date for daily repeat: %v", nextDate)
		return nextDate.Format("20060102"), nil

	case repeat == "y":
		nextDate := parsedDate.AddDate(1, 0, 0)
		if parsedDate.Month() == time.February && parsedDate.Day() == 29 && !isLeap(nextDate.Year()) {
			nextDate = nextDate.AddDate(0, 0, 1)
		}
		log.Printf("Next date for yearly repeat: %v", nextDate)
		return nextDate.Format("20060102"), nil

	default:
		return "", fmt.Errorf("unsupported repeat rule")
	}
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 != 0)
}
