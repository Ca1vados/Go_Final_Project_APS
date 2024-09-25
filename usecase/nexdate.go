package usecase

import (
	"fmt"
	"time"

	cnst "github.com/siavoid/task-manager/usecase/constants"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", nil
	}

	dataTime, err := time.Parse(cnst.DateFormat, date)
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
				if newDate.After(now) || newDate.Format(cnst.DateFormat) == dataTime.Format(cnst.DateFormat) {
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
			if newDate.After(now) || newDate.Format(cnst.DateFormat) == dataTime.Format(cnst.DateFormat) {
				break
			}
			dataTime = newDate
		}
	} else {
		return "", fmt.Errorf("incorrect repeat format: \"%s\" ", repeat)
	}

	return newDate.Format(cnst.DateFormat), nil
}
