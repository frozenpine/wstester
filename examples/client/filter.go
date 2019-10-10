package main

import (
	"context"

	linq "github.com/ahmetb/go-linq"
	"github.com/frozenpine/ngerest"
	"github.com/frozenpine/wstester/models"
	flag "github.com/spf13/pflag"
)

var (
	filterStr string
)

func init() {
	flag.StringVar(&filterStr, "filter", "", "Filter for result.")
}

func filter(ctx context.Context, table string, ch <-chan models.TableResponse) <-chan models.TableResponse {
	filterChan := make(chan models.TableResponse)

	var filterFunc = func(rsp models.TableResponse) models.TableResponse {
		return rsp
	}

	switch table {
	case "trade":
		filterFunc = func(rsp models.TableResponse) models.TableResponse {
			return rsp
		}
	case "instrument":
		filterFunc = func(rsp models.TableResponse) models.TableResponse {
			insRsp := rsp.(*models.InstrumentResponse)

			var filted []*ngerest.Instrument

			linq.From(insRsp.Data).WhereT(func(c *ngerest.Instrument) bool {
				return c.MarkPrice > 0 && c.FairPrice > 0
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
	case "orderBookL2":
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