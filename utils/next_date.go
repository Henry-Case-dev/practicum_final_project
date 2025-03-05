package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату задачи.
// now – переданное время (из параметра now запроса), dateStr – исходная дата задачи,
// repeat поддерживает два формата: "d N" – для ежедневного повторения и "y" – для ежегодного.
func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	log.Printf("DEBUG (NextDate): now=%s, dateStr=%s, repeat=%s", now.Format("20060102"), dateStr, repeat)
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
		// Если исходная дата равна now (как должно быть для повторяющихся задач), просто прибавляем интервал.
		if date.Equal(now) {
			next := date.AddDate(0, 0, days)
			log.Printf("DEBUG (NextDate): equal case, next=%s", next.Format("20060102"))
			return next.Format("20060102"), nil
		}
		// Иначе начинаем с date + days и повторяем прибавление, пока next не станет строго больше now.
		next := date.AddDate(0, 0, days)
		log.Printf("DEBUG (NextDate): initial next=%s", next.Format("20060102"))
		iteration := 0
		for !next.After(now) {
			iteration++
			next = next.AddDate(0, 0, days)
			log.Printf("DEBUG (NextDate): iteration %d, updated next=%s", iteration, next.Format("20060102"))
		}
		log.Printf("DEBUG (NextDate): loop iterations=%d, final next=%s", iteration, next.Format("20060102"))
		return next.Format("20060102"), nil

	} else if repeat == "y" {
		var targetYear int
		if date.Year() < now.Year() {
			targetYear = now.Year()
		} else {
			targetYear = date.Year() + 1
		}
		next := time.Date(targetYear, date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		log.Printf("DEBUG (NextDate): rule y: initial next=%s", next.Format("20060102"))
		if !next.After(now) {
			next = next.AddDate(1, 0, 0)
			log.Printf("DEBUG (NextDate): rule y: adjusted next=%s", next.Format("20060102"))
		}
		return next.Format("20060102"), nil

	} else {
		return "", fmt.Errorf("неподдерживаемый формат повторения: %s", repeat)
	}
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
