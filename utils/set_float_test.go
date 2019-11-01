package utils

import (
	"testing"
)

func TestFloat64Set(t *testing.T) {
	prices := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	set := NewFloat64Set(prices)
	if set.Len() > 10 {
		t.Fatal("unique failed")
	}

	sub := set.Sub(NewFloat64Set([]float64{1, 2, 3})).(Float64Set)
	if sub.Len() != 7 {
		t.Fatal("sub failed.")
	}
	if sub.Exist(1) || sub.Exist(2) || sub.Exist(3) {
		t.Fatal("sub failed")
	}

	if set.Len() != 10 {
		t.Fatal("failed")
	}
}
