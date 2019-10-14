package utils

import (
	"testing"

	"github.com/frozenpine/ngerest"
)

func TestGetField(t *testing.T) {
	td := ngerest.Trade{
		Symbol: "XBTUSD",
		Size:   100,
		Side:   "Buy",
		Price:  10000,
	}

	t.Log(GetFields(td, "Symbol", "Side", "Size", "Price"))
}

func TestParse(t *testing.T) {
	RegisterTableModel("trade", new(ngerest.Trade))
	RegisterTableModel("instrument", new(ngerest.Instrument))
	RegisterTableModel("orderBookL2", new(ngerest.OrderBookL2))

	sql := `select Symbol, Price, Size as volume, Timestamp from trade where Price > 0.0 or Symbol = 'XBTUSD' UNION select FairPrice, Symbol from instrument union select * from orderBookL2`

	td := ngerest.Trade{
		Symbol: "XBTUSD",
		Price:  0,
	}

	filters, err := ParseSQL(sql)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(filters["trade"].GetFilter()([]*ngerest.Trade{&td}))
	t.Log(filters["instrument"].GetFilter()([]*ngerest.Instrument{
		&ngerest.Instrument{MarkPrice: 10000, FairPrice: 10000},
	}))
	t.Log(filters["orderBookL2"].GetFilter()([]*ngerest.OrderBookL2{
		&ngerest.OrderBookL2{
			Price:  9203,
			Side:   "Sell",
			Symbol: "XBTUSD",
			Size:   100,
		},
	}))
}
