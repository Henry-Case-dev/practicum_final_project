package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// AppTimeZone - фиксированный часовой пояс для работы с датами.
// Используется для обеспечения согласованности независимо от системного часового пояса.
var AppTimeZone = time.UTC

// StartOfDay возвращает дату с нулевым временем (00:00:00) в указанном часовом поясе
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, AppTimeZone)
}

// ParseDateString преобразует строку даты в формате "YYYYMMDD" в time.Time
func ParseDateString(dateStr string) (time.Time, error) {
	// Парсим дату без учета часового пояса
	t, err := time.Parse("20060102", dateStr)
	if err != nil {
		return time.Time{}, err
	}

	// Извлекаем компоненты даты и создаем новый объект time.Time с нулевым временем в AppTimeZone
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, AppTimeZone), nil
}

// NextDate вычисляет следующую дату задачи.
// now – переданное время (из параметра now запроса), dateStr – исходная дата задачи,
// repeat поддерживает два формата: "d N" – для ежедневного повторения и "y" – для ежегодного.
func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	log.Printf("DEBUG (NextDate): now=%s, dateStr=%s, repeat=%s", now.Format("20060102"), dateStr, repeat)
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	// Устанавливаем начало дня для now
	now = StartOfDay(now)

	// Парсим и обрабатываем дату
	date, err := ParseDateString(dateStr)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

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
		next := time.Date(targetYear, date.Month(), date.Day(), 0, 0, 0, 0, AppTimeZone)
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
