package mock

import (
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
)

// Trade mock trade response
func Trade(cache utils.Cache) {
	ticker := time.NewTicker(time.Second)

	var (
		lastPrice         float64
		lastTickDirection string
		sides                     = [2]string{"Buy", "Sell"}
		sizeMax                   = float32(1000.0)
		priceMax                  = 9900.0
		priceMin                  = 8100.0
		hisMaxRate        float64 = 0.0
	)

	for {
		<-ticker.C

		start := time.Now()
		rand.Seed(start.UnixNano())
		count := rand.Intn(1000)

		mockTrad := models.TradeResponse{}
		mockTrad.Table = "trade"
		mockTrad.Action = "insert"

		for i := 0; i < count; i++ {

			choice := rand.Intn(1000)

			tickDirection := ""

			price := priceMin + float64(choice)*0.5
			if price > priceMax {
				price = priceMax
			}
			switch {
			case price > lastPrice:
				tickDirection = "PlusTick"
			case price == lastPrice:
				if strings.Contains(lastTickDirection, "Zero") {
					tickDirection = lastTickDirection
				} else {
					tickDirection = "Zero" + lastTickDirection
				}
			case price < lastPrice:
				tickDirection = "MinusTick"
			}

			lastTickDirection = tickDirection

			size := float32(choice)
			if size < 1 {
				size = 1
			} else if size > sizeMax {
				size = sizeMax
			}

			lastPrice = price

			td := ngerest.Trade{
				Symbol:        "XBTUSD",
				Side:          sides[choice%2],
				Size:          size,
				Price:         price,
				Timestamp:     ngerest.NGETime(time.Now()),
				TickDirection: tickDirection,
			}

			lastPrice = td.Price

			mockTrad.Data = append(mockTrad.Data, &td)
		}

		cache.Append(utils.NewCacheInput(&mockTrad))

		elasped := time.Now().Sub(start).Nanoseconds()
		rate := float64(count) * 1000.0 / float64(elasped/1000)
		if rate > hisMaxRate && !math.IsInf(rate, 1) {
			hisMaxRate = rate
		}

		log.Printf("Mock trade send rate: %.2f rps, history max rate: %.2f rps\n", rate, hisMaxRate)
	}
}
