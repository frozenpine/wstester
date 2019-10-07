package utils

import "sort"

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
