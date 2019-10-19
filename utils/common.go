package utils

import (
	"math"
)

// MaxInt return max int num in args
func MaxInt(numbers ...int) int {
	max := math.MinInt64

	for _, num := range numbers {
		if num > max {
			max = num
		}
	}

	return max
}

// MinInt return min int num in args
func MinInt(numbers ...int) int {
	min := math.MaxInt64

	for _, num := range numbers {
		if num < min {
			min = num
		}
	}

	return min
}
