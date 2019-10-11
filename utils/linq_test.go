package utils

import (
	"testing"

	"github.com/frozenpine/ngerest"
	"github.com/xwb1989/sqlparser"
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
	sql := "select a.Symbol, b.Price from trade a, instrument b"
	// sql := "select Symbol, Price from trade"

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		t.Fatal(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		tblDefine := parseTables(stmt)
		t.Log(tblDefine)

		colDefine := parseColumns(tblDefine, stmt)
		t.Log(colDefine)
	default:
		t.Fatal("invalid sql")
	}
}
