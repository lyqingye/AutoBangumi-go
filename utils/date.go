package utils

import "time"

func ParseDate(str string) (time.Time, error) {
	return time.Parse("2006-01-02", str)
}

func SmartParseDate(str string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	var t time.Time
	var err error

	for _, layout := range layouts {
		t, err = time.ParseInLocation(layout, str, time.Local)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}

func FindCloseTime(times []time.Time, target time.Time) int {
	if len(times) == 0 {
		panic("empty array")
	}
	closestIndex := 0
	minDiff := target.Sub(times[0]).Abs()

	for i := 1; i < len(times); i++ {
		diff := target.Sub(times[i]).Abs()
		if diff < minDiff {
			minDiff = diff
			closestIndex = i
		}
	}

	return closestIndex
}
