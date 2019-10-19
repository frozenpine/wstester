package utils

import (
	"math"
)

// MaxInt return max int
func MaxInt(i, j int) int {
	if i >= j {
		return i
	}
	return j
}

// MinInt return min int
func MinInt(i, j int) int {
	if i <= j {
		return i
	}
	return j
}

// MaxInts return max int num in args
func MaxInts(numbers ...int) int {
	max := math.MinInt64

	for _, num := range numbers {
		if num > max {
			max = num
		}
	}

	return max
}

// MinInts return min int num in args
func MinInts(numbers ...int) int {
	min := math.MaxInt64

	for _, num := range numbers {
		if num < min {
			min = num
		}
	}

	return min
}
