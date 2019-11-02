package mock

import (
	"context"
	"time"

	"github.com/frozenpine/wstester/client"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
	"github.com/frozenpine/wstester/utils/log"
)

// Upstream get mbl|trade|instrument response from upstream www.btcmex.com
func Upstream(caches map[string]utils.Cache) {
	for {
		cfg := client.NewConfig()
		cfg.DisableCache()
		ins := client.NewClient(cfg)

		var topics []string
		for topic := range caches {
			topics = append(topics, topic)
		}
		ins.Subscribe(topics...)

		ctx, cancelFn := context.WithCancel(context.Background())

		err := ins.Connect(ctx)
		if err != nil {
			log.Error(err)

			time.Sleep(time.Second * 3)

			continue
		}

		go func() {
			mblChan := ins.GetResponse("orderBookL2")
			tdChan := ins.GetResponse("trade")
			insChan := ins.GetResponse("instrument")

			var (
				rsp   models.TableResponse
				topic string
				ok    bool
			)

			for {
				select {
				case <-ctx.Done():
					return
				case rsp, ok = <-tdChan:
					topic = "trade"
				case rsp, ok = <-insChan:
					topic = "instrument"
				case rsp, ok = <-mblChan:
					topic = "orderBookL2"
				}

				if !ok {
					cancelFn()
					return
				}

				if rsp == nil {
					continue
				}

				caches[topic].Append(utils.NewCacheInput(rsp))
			}
		}()

		<-ins.Closed()

		log.Warn("Mock MBL upstream closed.")

		cancelFn()

		<-time.After(time.Second * 5)
	}
}
