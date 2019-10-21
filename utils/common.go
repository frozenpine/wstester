package utils

import (
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

// PriceSort insert price in price list
func PriceSort(priceList []float64, price float64, desc bool) (int, []float64) {
	originLen := len(priceList)

	midIdx := originLen / 2

	if desc {
		if price > priceList[0] {
			rtn := make([]float64, originLen+1)
			rtn[0] = price
			for idx, price := range priceList {
				rtn[idx+1] = price
			}
			return 0, rtn
		}

		if price < priceList[originLen-1] {
			return len(priceList), append(priceList, price)
		}

		for {
			if price < priceList[midIdx] {
				if price > priceList[midIdx+1] {
					// TODO: 修正right部分数据被覆盖的问题
					return midIdx + 1, append(append(priceList[0:midIdx+1], price), priceList[midIdx+2:originLen]...)
				}

				newIdx := midIdx / 2

				if midIdx == newIdx {
					midIdx++
				} else {
					midIdx = newIdx
				}

				if midIdx >= originLen {
					break
				}
			} else {
				if price < priceList[midIdx-1] {
					return midIdx, append(append(priceList[0:midIdx], price), priceList[midIdx+1:originLen]...)
				}

				newIdx := (originLen - midIdx) / 2

				if midIdx == newIdx {
					midIdx--
				} else {
					midIdx = newIdx
				}

				if midIdx <= 0 {
					break
				}
			}
		}

		return -1, nil
	}

	if price > priceList[originLen-1] {
		return originLen, append(priceList, price)
	}

	if price < priceList[0] {
		rtn := make([]float64, originLen+1)
		rtn[0] = price
		for idx, price := range priceList {
			rtn[idx+1] = price
		}
		return 0, rtn
	}

	for {
		if price > priceList[midIdx] {
			if price < priceList[midIdx+1] {
				return midIdx + 1, append(append(priceList[0:midIdx+1], price), priceList[midIdx+2:]...)
			}

			newIdx := (originLen - midIdx) / 2

			if midIdx == newIdx {
				midIdx++
			} else {
				midIdx = newIdx
			}

			if midIdx >= originLen {
				break
			}
		} else {
			if price > priceList[midIdx-1] {
				return midIdx, append(append(priceList[0:midIdx], price), priceList[midIdx+1:originLen]...)
			}

			newIdx := midIdx / 2

			if midIdx == newIdx {
				midIdx--
			} else {
				midIdx = newIdx
			}

			if midIdx <= 0 {
				break
			}
		}
	}

	return -1, nil
}
