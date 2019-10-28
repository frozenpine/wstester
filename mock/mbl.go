package mock

import (
	"context"
	"time"

	"github.com/frozenpine/wstester/client"
	"github.com/frozenpine/wstester/utils"
)

// MBL mock mbl response
func MBL(cache utils.Cache) {
	for {
		cfg := client.NewConfig()
		ins := client.NewClient(cfg)
		ins.Subscribe("orderBookL2")

		ctx, cancelFn := context.WithCancel(context.Background())

		ins.Connect(ctx)

		go func() {
			mblChan := ins.GetResponse("orderBookL2")

			for {
				select {
				case <-ctx.Done():
					return
				case mbl, ok := <-mblChan:
					if !ok {
						cancelFn()
						return
					}

					if mbl == nil {
						continue
					}

					cache.Append(utils.NewCacheInput(mbl))
				}
			}
		}()

		<-ins.Closed()

		cancelFn()

		<-time.After(time.Second * 5)
	}
}
