package utils

import (
	"testing"

	"github.com/frozenpine/ngerest"
)

func TestSnapshot(t *testing.T) {
	cache := MBLCache{}
	cache.initCache()

	t.Log(cache.snapshot(0))

	for _, ask := range []float64{9995, 9996, 9997, 9998, 9999, 10000, 10001} {
		cache.handleInsert(&ngerest.OrderBookL2{
			Price: float64(ask),
			Size:  float32(ask),
			Side:  "Sell",
		})
	}
	for _, bid := range []float64{9990, 9991, 9992, 9993, 9994} {
		cache.handleInsert(&ngerest.OrderBookL2{
			Price: float64(bid),
			Size:  float32(bid),
			Side:  "Buy",
		})
	}

	t.Log(cache.snapshot(3))
	t.Log(cache.askPrices)
	t.Log(cache.bidPrices)

	t.Log(cache.snapshot(6))
	t.Log(cache.askPrices)
	t.Log(cache.bidPrices)

	t.Log(cache.snapshot(9))
	t.Log(cache.askPrices)
	t.Log(cache.bidPrices)
}

func BenchmarkInsertBuy(b *testing.B) {
	cache := MBLCache{}
	cache.initCache()

	for i := 0; i < b.N; i++ {
		if _, err := cache.handleInsert(&ngerest.OrderBookL2{
			Price: float64(i),
			Size:  float32(i),
			Side:  "Buy",
		}); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkInsertSell(b *testing.B) {
	cache := MBLCache{}
	cache.initCache()

	for i := 0; i < b.N; i++ {
		if _, err := cache.handleInsert(&ngerest.OrderBookL2{
			Price: float64(b.N - i),
			Size:  float32(i),
			Side:  "Sell",
		}); err != nil {
			b.Error(err)
		}
	}
}
