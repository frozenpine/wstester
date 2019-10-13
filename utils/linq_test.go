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

	sql := `select Symbol, Price, Size as volume from trade where Price > 1.0 or Symbol = 'XBTUSD' UNION select MarkPrice, FairPrice from instrument union select * from orderBookL2`

	td := ngerest.Trade{
		Symbol: "XBTUSD",
		Price:  0,
	}

	fn, err := ParseSQL(sql)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(fn["trade"]([]*ngerest.Trade{&td}))
}
