package helper

import "math"

func CalculatePercentageChange(current, previous int64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0 // No change if both are zero
		}
		return 100 // 100% increase if there were none before but some now
	}
	return math.Round(float64(current-previous)/float64(previous)*10000) / 100
}
