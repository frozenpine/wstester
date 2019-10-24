package utils

import (
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestPriceAdd(t *testing.T) {
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

func TestReadAdd(t *testing.T) {
	testList := sort.Float64Slice{30000, 9900, 8300, 8160, 7559.5, 7557, 7554, 7553, 7552.5, 7551, 7550, 7548.5, 7548,
		7546.5, 7545, 7544.5, 7542, 7540.5, 7540, 7539, 7536.5, 7536, 7533.5, 7533, 7532.5, 7530, 7528.5, 7527, 7524.5,
		7524, 7521, 7520.5, 7518, 7516.5, 7515, 7514, 7512.5, 7512, 7509, 7508.5, 7486, 7485, 7484, 7483, 7482, 7481,
		7480, 7479, 7478, 7477, 7476, 7475, 7472, 7471, 7469, 7468.5, 7468, 7467.5, 7467, 7466, 7465, 7464.5, 7464,
		7462.5, 7462, 7461, 7460, 7459.5, 7459, 7458.5, 7458, 7457.5, 7457}

	var idx int

	idx, testList = PriceAdd(testList, 7463.5, true)

	t.Log(idx, testList)
}

func TestPriceRemove(t *testing.T) {
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
}

func TestPriceSearch(t *testing.T) {
	bids := sort.Float64Slice{1, 3, 5, 7, 9}
	t.Log(PriceSearch(bids, 10, false))
	t.Log(PriceSearch(bids, 0, false))
	t.Log(PriceSearch(bids, 4, false))
	t.Log(PriceSearch(bids, 9, false))
	t.Log(PriceSearch(bids, 1, false))

	asks := sort.Float64Slice{9, 7, 5, 3, 1}
	t.Log(PriceSearch(asks, 10, true))
	t.Log(PriceSearch(asks, 0, true))
	t.Log(PriceSearch(asks, 4, true))
	t.Log(PriceSearch(asks, 2, true))
}
