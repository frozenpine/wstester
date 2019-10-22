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
		t.Error("mid left failed.")
	} else {
		t.Log(result)
	}

	idx, result = PriceSort(result, 43, false)
	if idx != 6 {
		t.Error("mid right failed.")
	} else {
		t.Log(result)
	}

	bids := []float64{50, 40, 30, 20, 10}

	idx, result = PriceSort(bids, 78, true)
	if idx != 0 {
		t.Error("head failed.")
	} else {
		t.Log(result)
	}

	idx, result = PriceSort(result, 2, true)
	if idx != 6 {
		t.Error("tail failed.")
	} else {
		t.Log(result)
	}

	idx, result = PriceSort(result, 46, true)
	if idx != 2 {
		t.Error("mid left failed.")
	} else {
		t.Log(result)
	}

	idx, result = PriceSort(result, 6, true)
	if idx != 7 {
		t.Error("mid right failed.")
	} else {
		t.Log(result)
	}
}
