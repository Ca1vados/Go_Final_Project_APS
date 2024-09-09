package usecase

import (
	"fmt"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", nil
	}

	dataTime, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("incorrect date: \"%s\" ", date)
	}
	var newDate time.Time
	var datePart string
	var n int
	fmt.Sscanf(repeat, "%s %d", &datePart, &n)
	if datePart == "d" {
		if n > 0 && n <= 400 {
			for {
				newDate = dataTime.AddDate(0, 0, n)
				if newDate.After(now) || newDate.Format("20060102") == dataTime.Format("20060102") {
					break
				}
				dataTime = newDate
			}
		} else {
			return "", fmt.Errorf("incorrect repeat format: \"%s\" ", repeat)
		}
	} else if datePart == "y" {
		for {
			newDate = dataTime.AddDate(1, 0, 0)
			if newDate.After(now) || newDate.Format("20060102") == dataTime.Format("20060102") {
				break
			}
			dataTime = newDate
		}
	} else {
		return "", fmt.Errorf("incorrect repeat format: \"%s\" ", repeat)
	}

	return newDate.Format("20060102"), nil
}
