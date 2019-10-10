package main

import (
	"bytes"
	"context"
	"text/template"

	linq "github.com/ahmetb/go-linq"
	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	flag "github.com/spf13/pflag"
)

var (
	filterList []string
)

func init() {
	flag.StringSliceVar(&filterList, "filter", []string{}, "Filter for result.")
}

func makeTradeFilter(tpl *template.Template) func(models.TableResponse) models.TableResponse {
	return func(rsp models.TableResponse) models.TableResponse {
		tdRsp, ok := rsp.(*models.TradeResponse)
		if !ok {
			panic("response is not Trade.")
		}

		var filted []*ngerest.Trade

		linq.From(tdRsp.Data).WhereT(func(c *ngerest.Trade) bool {
			buf := bytes.Buffer{}

			if err := tpl.Execute(&buf, c); err != nil {
				panic(err)
			}

			return buf.Len() > 0
		}).ToSlice(&filted)

		if len(filted) > 0 {
			if len(filted) == len(tdRsp.Data) {
				return rsp
			}

			newRsp := models.TradeResponse{}

			newRsp.Table = tdRsp.Table
			newRsp.Action = tdRsp.Action
			newRsp.Keys = tdRsp.Keys
			newRsp.Types = tdRsp.Types
			newRsp.ForeignKeys = tdRsp.ForeignKeys
			newRsp.Attributes = tdRsp.Attributes
			newRsp.Data = filted

			return &newRsp
		}

		return nil
	}
}

func makeInstrumentFilter(tpl *template.Template) func(models.TableResponse) models.TableResponse {
	return func(rsp models.TableResponse) models.TableResponse {
		insRsp := rsp.(*models.InstrumentResponse)

		var filted []*ngerest.Instrument

		linq.From(insRsp.Data).WhereT(func(c *ngerest.Instrument) bool {
			buf := bytes.Buffer{}

			if err := tpl.Execute(&buf, c); err != nil {
				panic(err)
			}

			return buf.Len() > 0
		}).ToSlice(&filted)

		if len(filted) > 0 {
			if len(filted) == len(insRsp.Data) {
				return rsp
			}

			newRsp := models.InstrumentResponse{}

			newRsp.Table = insRsp.Table
			newRsp.Action = insRsp.Action
			newRsp.Keys = insRsp.Keys
			newRsp.Types = insRsp.Types
			newRsp.ForeignKeys = insRsp.ForeignKeys
			newRsp.Attributes = insRsp.Attributes
			newRsp.Data = filted

			return &newRsp
		}

		return nil
	}
}

func makeMBLFilter(tpl *template.Template) func(models.TableResponse) models.TableResponse {
	return func(rsp models.TableResponse) models.TableResponse {
		mblRsp := rsp.(*models.MBLResponse)

		var filted []*ngerest.OrderBookL2

		linq.From(mblRsp.Data).WhereT(func(c *ngerest.OrderBookL2) bool {
			buf := bytes.Buffer{}

			if err := tpl.Execute(&buf, c); err != nil {
				panic(err)
			}

			return buf.Len() > 0
		}).ToSlice(&filted)

		if len(filted) > 0 {
			if len(filted) == len(mblRsp.Data) {
				return rsp
			}

			newRsp := models.MBLResponse{}

			newRsp.Table = mblRsp.Table
			newRsp.Action = mblRsp.Action
			newRsp.Keys = mblRsp.Keys
			newRsp.Types = mblRsp.Types
			newRsp.ForeignKeys = mblRsp.ForeignKeys
			newRsp.Attributes = mblRsp.Attributes
			newRsp.Data = filted

			return &newRsp
		}

		return nil
	}
}

func makeFilterMap() map[string]func(models.TableResponse) models.TableResponse {
	tableLen := len(tables)

	filters := make(map[string]func(models.TableResponse) models.TableResponse)

	for idx, filter := range filterList {
		tpl, err := template.New("filter").Parse(filter)

		if err != nil {
			panic(err)
		}

		if idx >= tableLen {
			break
		}

		tableName := tables[idx]

		switch tableName {
		case "trade":
			filters[tableName] = makeTradeFilter(tpl)
		case "instrument":
			filters[tableName] = makeInstrumentFilter(tpl)
		case "orderBookL2":
			filters[tableName] = makeMBLFilter(tpl)
		}
	}

	return filters
}

func filter(ctx context.Context, table string, ch <-chan models.TableResponse) <-chan models.TableResponse {
	filterChan := make(chan models.TableResponse)

	filterFunc, exist := makeFilterMap()[table]
	if !exist {
		filterFunc = func(rsp models.TableResponse) models.TableResponse {
			return rsp
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case rsp := <-ch:
				if rsp == nil {
					return
				}

				if rsp = filterFunc(rsp); rsp != nil {
					filterChan <- rsp
				}
			}
		}
	}()

	return filterChan
}
