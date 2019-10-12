package utils

import (
	"math"
	"sort"
)

// RangeSlice trim elements specified by a index slice
func RangeSlice(src []interface{}, sliced []int) []interface{} {
	if sliced != nil && len(sliced) > 0 {
		sort.Sort(sort.IntSlice(sliced))

		invalid := 0

		for count, idx := range sliced {
			trim := idx - count + invalid

			if trim < 0 {
				invalid++
				continue
			}

			if trim >= len(src) {
				break
			}

			src = append(src[:trim], src[trim+1:]...)
		}
	}

	return src
}

// Slice trim a element specified by index
func Slice(src []interface{}, idx int) []interface{} {
	if idx >= 0 && idx < len(src) {
		src = append(src[:idx], src[idx+1:]...)
	}

	return src
}

// MaxInt return max int num in args
func MaxInt(numbers ...int) int {
	max := 0

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
