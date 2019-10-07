package utils

import "testing"

func TestRangeSlice(t *testing.T) {
	src := []string{"a", "b", "c", "d", "e", "f", "g"}
	slice := []int{-1, -3, 3, 8}

	tmpSlice := make([]interface{}, len(src))
	for idx, ele := range src {
		tmpSlice[idx] = ele
	}

	tmpSlice = RangeSlice(tmpSlice, slice)

	t.Log(tmpSlice)
}

func TestSlice(t *testing.T) {
	src := []string{"a", "b", "c", "d", "e", "f", "g"}

	t.Log(src[len(src):])

	tmpSlice := make([]interface{}, len(src))
	for idx, ele := range src {
		tmpSlice[idx] = ele
	}

	t.Log(Slice(tmpSlice, 7))
}
