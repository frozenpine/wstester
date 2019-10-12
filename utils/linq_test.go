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
	// sql := "select a.Symbol, b.Price from trade a, instrument b where a.Size > 100 and (b.MarkPrice > 0 or b.FairPrice > 0)"
	sql := `select Symbol, Price from trade where Price > 1.0 or Symbol = 'XBTUSD'`

	td := ngerest.Trade{
		Symbol: "XBTUSD",
		Price:  0,
	}

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		t.Fatal(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		tblDefine, err := parseTables(stmt)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(tblDefine)

		colDefine, err := parseColumns(tblDefine, stmt)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(colDefine)

		t.Log(sqlparser.String(stmt.Where.Expr))

		conditionFn, err := conditionParser(stmt.Where.Expr)
		if err != nil {
			t.Fatal(err)
		}

		t.Log(conditionFn(&td))
	default:
		t.Fatal("unsupported statement: " + sqlparser.String(stmt))
	}
}
