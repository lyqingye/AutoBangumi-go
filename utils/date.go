package utils

import "time"

func ParseDate(str string) (time.Time, error) {
	return time.Parse("2006-01-02", str)
}

func FindCloseTime(times []time.Time, target time.Time) int {
	if len(times) == 0 {
		panic("empty array")
	}
	closestIndex := 0
	minDiff := target.Sub(times[0])

	for i := 1; i < len(times); i++ {
		diff := target.Sub(times[i])
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			closestIndex = i
		}
	}

	return closestIndex
}
