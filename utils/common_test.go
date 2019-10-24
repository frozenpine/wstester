package utils

import (
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestPriceSort(t *testing.T) {
	listLength := 100
	basePrice := 8000.0
	priceTick := 0.5

	for i := 0; i < listLength; i++ {
		var (
			priceList, acs, desc sort.Float64Slice
			idx                  int
		)

		for i := 0; i < listLength; i++ {
			priceList = append(priceList, basePrice+priceTick*float64(i))
		}

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(listLength, func(i, j int) { priceList[i], priceList[j] = priceList[j], priceList[i] })

		for _, price := range priceList {
			originLen := acs.Len()
			idx, acs = PriceAdd(acs, price, false)
			addLen := acs.Len()

			if addLen-originLen != 1 {
				t.Fatal("add price failed:", acs, price, idx, originLen, addLen)
			}

			searchIdx := acs.Search(price)

			if idx != searchIdx {
				t.Fatal("ASC sort failed:", acs, price, idx, searchIdx)
			}
		}

		for _, price := range priceList {
			originLen := desc.Len()
			idx, desc = PriceAdd(desc, price, false)
			addLen := desc.Len()

			if addLen-originLen != 1 {
				t.Fatal("add price failed:", desc, price, idx, originLen, addLen)
			}

			searchIdx := desc.Search(price)

			if idx != searchIdx {
				t.Fatal("DESC sort failed:", desc, price, idx, searchIdx)
			}
		}
	}
}

func TestPriceRemove(t *testing.T) {
	listLength := 100

	basePrice := 8000.0
	priceTick := 0.5

	var (
		priceList, acs, desc sort.Float64Slice
		idx                  int
	)

	for i := 0; i < listLength; i++ {
		priceList = append(priceList, basePrice+priceTick*float64(i))
	}

	acs = append(sort.Float64Slice{}, priceList...)
	desc = append(sort.Float64Slice{}, priceList...)
	ReverseFloat64Slice(desc)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(listLength, func(i, j int) { priceList[i], priceList[j] = priceList[j], priceList[i] })

	for _, price := range priceList {
		originLen := acs.Len()
		idx, acs = PriceRemove(acs, price, false)
		removeLen := acs.Len()

		if originLen-removeLen != 1 {
			t.Fatal("remove more than 1 price:", acs, price, idx, originLen, removeLen)
		}

		if removeLen < 1 {
			break
		}

		if idx == 0 {
			if price >= acs[0] {
				t.Fatal("remove wrong price:", acs, price, acs[0], idx)
			}
		} else if idx == originLen-1 {
			if price <= acs[removeLen-1] {
				t.Fatal("remove wrong price:", acs, price, acs[removeLen-1], idx)
			}
		} else if acs[idx-1] >= price || acs[idx] <= price {
			t.Fatal("remove wrong price:", acs, price, acs[idx-1], acs[idx], idx)
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(listLength, func(i, j int) { priceList[i], priceList[j] = priceList[j], priceList[i] })

	for _, price := range priceList {
		originLen := desc.Len()
		idx, desc = PriceRemove(desc, price, true)
		removeLen := desc.Len()

		if originLen-removeLen != 1 {
			t.Fatal("remove more than 1 price:", desc, price, idx, originLen, removeLen)
		}

		if removeLen < 1 {
			break
		}

		if idx == 0 {
			if price <= desc[0] {
				t.Fatal("remove wrong price:", desc, price, desc[0], idx)
			}
		} else if idx == originLen-1 {
			if price >= desc[removeLen-1] {
				t.Fatal("remove wrong price:", desc, price, idx)
			}
		} else if desc[idx-1] <= price || desc[idx] >= price {
			t.Fatal("remove wrong price:", desc, price, desc[idx-1], desc[idx], idx)
		}
	}
}

func TestPriceSearch(t *testing.T) {
	asks := sort.Float64Slice{1, 3, 5, 7, 9}
	t.Log(PriceSearch(asks, 10, false))
	t.Log(PriceSearch(asks, 0, false))
	t.Log(PriceSearch(asks, 4, false))

	bids := sort.Float64Slice{9, 7, 5, 3, 1}
	t.Log(PriceSearch(bids, 10, true))
	t.Log(PriceSearch(bids, 0, true))
	t.Log(PriceSearch(bids, 4, true))
}
