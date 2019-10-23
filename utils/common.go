package utils

import (
	"log"
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

// PriceSort insert price in price list, origin price list must be sorted, and has unique price
func PriceSort(priceList []float64, price float64, reverse bool) (int, []float64) {
	log.Println("price sort begin:", priceList, price, reverse)
	defer log.Println("price sort end:", priceList)

	originLen := len(priceList)

	if originLen < 1 {
		return 0, append(priceList, price)
	}

	start := 0
	end := originLen - 1

	if reverse {
		if price > priceList[0] {
			rtn := make([]float64, originLen+1)

			rtn[0] = price
			copy(rtn[1:], priceList)

			return 0, rtn
		}

		if price < priceList[originLen-1] {
			return len(priceList), append(priceList, price)
		}

		for {
			mid := (end-start)/2 + start

			if mid >= end {
				return len(priceList), append(priceList, price)
			}

			if mid <= start {
				rtn := make([]float64, originLen+1)

				rtn[0] = price
				copy(rtn[1:], priceList)

				return 0, rtn
			}

			if price < priceList[mid] {
				if price > priceList[mid+1] {
					rtn := make([]float64, originLen+1)

					copy(rtn, priceList[0:mid+1])
					copy(rtn[mid+2:], priceList[mid+1:])

					rtn[mid+1] = price

					return mid + 1, rtn
				}

				start = mid
			} else {
				if price < priceList[mid-1] {
					rtn := make([]float64, originLen+1)

					copy(rtn, priceList[0:mid])
					copy(rtn[mid+2:], priceList[mid+1:])

					rtn[mid] = price

					return mid, rtn
				}

				end = mid
			}
		}

		// return -1, nil
	}

	if price > priceList[end] {
		return end + 1, append(priceList, price)
	}

	if price < priceList[start] {
		rtn := make([]float64, originLen+1)

		rtn[0] = price
		copy(rtn[1:], priceList)

		return start, rtn
	}

	for {
		mid := (end-start)/2 + start

		if mid >= end {
			return len(priceList), append(priceList, price)
		}

		if mid <= start {
			rtn := make([]float64, originLen+1)

			rtn[0] = price
			copy(rtn[1:], priceList)

			return 0, rtn
		}

		if price > priceList[mid] {
			if price < priceList[mid+1] {
				rtn := make([]float64, originLen+1)

				copy(rtn, priceList[0:mid+1])
				copy(rtn[mid+2:], priceList[mid+1:])

				rtn[mid+1] = price

				return mid + 1, rtn
			}

			start = mid
		} else {
			if price > priceList[mid-1] {
				rtn := make([]float64, originLen+1)

				copy(rtn, priceList[0:mid])
				copy(rtn[mid+1:], priceList[mid:])

				rtn[mid] = price

				return mid, rtn
			}

			end = mid
		}
	}

	// return -1, nil
}
