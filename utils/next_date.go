package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату задачи.
// Для ежедневного повторения (формат "d N") прибавляет N дней к дате и, при необходимости,
// повторяет прибавление, пока полученная дата не станет строго больше указанного now.
// Для ежегодного повторения ("y"):
//   - Если год исходной даты меньше now.Year(), используется now.Year(), иначе – date.Year()+1.
//   - Если полученная дата не больше now, прибавляется ещё один год.
//
// В случае 29 февраля, если в выбранном году такого дня нет, можно (при необходимости) скорректировать дату.
func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

	now = now.Truncate(24 * time.Hour)
	date = date.Truncate(24 * time.Hour)

	if strings.HasPrefix(repeat, "d ") {
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", fmt.Errorf("некорректный формат повторения: %s", repeat)
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("недопустимое число дней: %s", parts[1])
		}
		// Начинаем с date + days, чтобы вернуть значение, строго больше заданной даты.
		next := date.AddDate(0, 0, days)
		for !next.After(now) {
			next = next.AddDate(0, 0, days)
		}
		return next.Format("20060102"), nil

	} else if repeat == "y" {
		var targetYear int
		if date.Year() < now.Year() {
			targetYear = now.Year()
		} else {
			targetYear = date.Year() + 1
		}
		next := time.Date(targetYear, date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		if !next.After(now) {
			next = next.AddDate(1, 0, 0)
		}
		return next.Format("20060102"), nil

	} else {
		return "", fmt.Errorf("неподдерживаемый формат повторения: %s", repeat)
	}
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
