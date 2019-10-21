package utils

import "testing"

func TestPriceSort(t *testing.T) {
	buys := []float64{10, 20, 30, 40, 50}

	idx, result := PriceSort(buys, 63, false)
	if idx != 5 {
		t.Error("tail failed.")
	} else {
		t.Log(result)
	}

	idx, result = PriceSort(result, 4, false)
	if idx != 0 {
		t.Error("head failed.")
	} else {
		t.Log(result)
	}

	idx, result = PriceSort(result, 21, false)
	if idx != 3 {
		t.Error("mid failed.")
	} else {
		t.Log(result)
	}
}
