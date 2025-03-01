package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату задачи.
// Для ежедневного повторения (формат "d N") функция всегда прибавляет интервал в N дней один раз,
// а затем, при необходимости, повторяет прибавление до тех пор, пока полученная дата не станет больше now.
// Для ежегодного повторения ("y"):
// • Если год исходной даты меньше текущего года, используется год now.
// • Иначе — используется исходный год плюс один.
// В случае 29 февраля, если в выбранном году такого дня нет, дата корректируется на 1 марта.
func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

	// Обрезаем время до начала дня
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
		// Всегда прибавляем минимум один шаг
		next := date.AddDate(0, 0, days)
		// Если полученная дата всё ещё не больше now, прибавляем интервал циклически
		for !next.After(now) {
			next = next.AddDate(0, 0, days)
		}
		return next.Format("20060102"), nil

	} else if repeat == "y" {
		var targetYear int
		// Если год исходной даты меньше текущего, используем now.Year()
		if date.Year() < now.Year() {
			targetYear = now.Year()
		} else {
			targetYear = date.Year() + 1
		}
		// Корректировка для 29 февраля
		if date.Month() == time.February && date.Day() == 29 && !isLeap(targetYear) {
			return time.Date(targetYear, time.March, 1, 0, 0, 0, 0, time.UTC).Format("20060102"), nil
		}
		next := time.Date(targetYear, date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		return next.Format("20060102"), nil

	} else {
		return "", fmt.Errorf("неподдерживаемый формат повторения: %s", repeat)
	}
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
