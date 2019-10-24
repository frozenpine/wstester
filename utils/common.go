package utils

import (
	"log"
	"math"
	"sort"
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

// PriceSearch search price in price list, -1 returned if price not found.
func PriceSearch(priceList sort.Float64Slice, price float64, reverse bool) (idx int) {
	// FIXME: debug output
	defer func() {
		if idx < 0 {
			log.Println(priceList, price, idx, reverse)
		}
	}()

	originLen := priceList.Len()

	idx = -1

	if reverse {
		idx = sort.Search(originLen, func(i int) bool { return priceList[i] <= price })
	} else {
		idx = sort.Search(originLen, func(i int) bool { return priceList[i] >= price })
	}

	if idx >= originLen || priceList[idx] != price {
		return -1
	}

	return
}

// PriceAdd insert price in price list, origin price list must be sorted, and has unique price
func PriceAdd(priceList sort.Float64Slice, price float64, reverse bool) (idx int, rtn sort.Float64Slice) {
	// FIXME: debug level log
	// defer func() {
	// 	log.Println("price sorted:", rtn, idx, price, reverse)
	// }()

	originLen := len(priceList)

	if originLen < 1 {
		idx = 0
		rtn = append(priceList, price)
		return
	}

	start := 0
	end := originLen - 1

	if reverse {
		if price > priceList[start] {
			idx = start
			rtn = append(sort.Float64Slice{price}, priceList...)

			return
		}

		if price < priceList[end] {
			idx = end + 1
			rtn = append(priceList, price)
			return
		}

		for {
			if (end - start) <= 1 {
				right := append(sort.Float64Slice{}, priceList[end:]...)

				idx = start + 1
				rtn = append(append(priceList[0:start+1], price), right...)

				return
			}

			mid := (end-start)/2 + start

			if price < priceList[mid] {
				if price > priceList[mid+1] {
					right := append(sort.Float64Slice{}, priceList[mid+1:]...)

					idx = mid + 1
					rtn = append(append(priceList[:mid+1], price), right...)

					return
				}

				start = mid
			} else {
				if price < priceList[mid-1] {
					right := append(sort.Float64Slice{}, priceList[mid:]...)

					idx = mid
					rtn = append(append(priceList[0:mid], price), right...)

					return
				}

				end = mid
			}
		}
	}

	if price > priceList[end] {
		idx = end + 1
		rtn = append(priceList, price)
		return
	}

	if price < priceList[start] {
		idx = start
		rtn = append(sort.Float64Slice{price}, priceList...)

		return
	}

	for {
		if (end - start) <= 1 {
			right := append(sort.Float64Slice{}, priceList[end:]...)

			idx = start + 1
			rtn = append(append(priceList[0:start+1], price), right...)

			return
		}

		mid := (end-start)/2 + start

		if price > priceList[mid] {
			if price < priceList[mid+1] {
				right := append(sort.Float64Slice{}, priceList[mid+1:]...)

				idx = mid + 1
				rtn = append(append(priceList[0:mid+1], price), right...)

				return
			}

			start = mid
		} else {
			if price > priceList[mid-1] {
				right := append(sort.Float64Slice{}, priceList[mid:]...)

				idx = mid
				rtn = append(append(priceList[0:mid], price), right...)

				return
			}

			end = mid
		}
	}
}

// PriceRemove remove price from price list
func PriceRemove(priceList sort.Float64Slice, price float64, reverse bool) (int, sort.Float64Slice) {
	idx := PriceSearch(priceList, price, reverse)

	if idx < 0 {
		return idx, priceList
	}

	lastIdx := priceList.Len() - 1

	if idx == 0 {
		return idx, priceList[1:]
	}

	if idx == lastIdx {
		return idx, priceList[0:idx]
	}

	right := append(sort.Float64Slice{}, priceList[idx+1:]...)
	copy(priceList[idx:], right)

	return idx, priceList[0:lastIdx]
}
