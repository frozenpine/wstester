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
	uuid "github.com/satori/go.uuid"
)

// Trade mock trade response
func Trade(cache utils.Cache) {
	var (
		lastPrice         float64
		lastTickDirection string
		sides                     = [2]string{"Buy", "Sell"}
		sizeMax                   = float32(1000.0)
		priceMax                  = 9900.0
		priceMin                  = 8100.0
		hisMaxRate        float64 = 0.0

		lastCount, lastElasped float64
		totalCount             int
		appStart               = time.Now()
	)

	for {
		start := time.Now()
		rand.Seed(start.UnixNano())
		time.Sleep(time.Second * time.Duration(rand.Intn(3)))
		count := rand.Intn(100) + 1

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
				TrdMatchID:    uuid.NewV4().String(),
			}

			lastPrice = td.Price

			mockTrad.Data = append(mockTrad.Data, &td)
		}

		cache.Append(utils.NewCacheInput(&mockTrad))

		end := time.Now()

		elasped := end.Sub(start).Nanoseconds()
		totalSpend := end.Sub(appStart).Seconds()

		rate := (float64(count) + lastCount) * 1000 * 1000 / (float64(elasped) + lastElasped)
		totalCount += count

		if math.IsInf(rate, 1) {
			lastCount += float64(count)
			lastElasped += float64(elasped)
		} else {
			if rate > hisMaxRate {
				hisMaxRate = rate
			}

			log.Printf(
				"Mock trade send[%d] rate: %.2f rps, history max rate: %.2f rps, Avg rate: %.2f\n",
				count+int(lastCount), rate, hisMaxRate, float64(totalCount)/float64(totalSpend),
			)

			lastCount = 0
			lastElasped = 0
		}
	}
}
