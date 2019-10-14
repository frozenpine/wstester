package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	"github.com/frozenpine/wstester/utils"
	flag "github.com/spf13/pflag"
)

var (
	sql     string
	filters map[string]*utils.TableDef

	topicMapper = map[string]interface{}{
		"trade":       new(ngerest.Trade),
		"instrument":  new(ngerest.Instrument),
		"orderBookL2": new(ngerest.OrderBookL2),
		"order":       new(ngerest.Order),
		"margin":      new(ngerest.Margin),
		"position":    new(ngerest.Position),
	}
)

func init() {
	flag.StringVar(&sql, "output", "", "SQL for output.")
}

func filter(ctx context.Context, table string, ch <-chan models.TableResponse) error {
	tableDef, exist := filters[table]
	if !exist {
		return errors.New("table definition not exists")
	}

	tableName := tableDef.GetAliasName()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case rsp := <-ch:
				if rsp == nil {
					return
				}

				for _, data := range tableDef.GetFilter()(rsp.GetData()) {
					result, _ := json.Marshal(data)

					log.Println(tableName, rsp.GetAction(), "<-", string(result))
				}
			}
		}
	}()

	return nil
}
