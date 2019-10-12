package utils

import (
	"strconv"
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

func parseComparison(compare *sqlparser.ComparisonExpr) func(interface{}) bool {
	left, ok := compare.Left.(*sqlparser.ColName)
	if !ok {
		panic("")
	}

	right, ok := compare.Right.(*sqlparser.SQLVal)
	if !ok {
		panic("")
	}

	return func(v interface{}) bool {
		switch compare.Operator {
		case sqlparser.EqualStr:
			switch right.Type {
			case sqlparser.IntVal:
				rightValue, _ := strconv.Atoi(string(right.Val))
				return GetFieldValue(v, left.Name.String()).(int) == rightValue
			case sqlparser.FloatVal:
				rightValue, _ := strconv.ParseFloat(string(right.Val), 64)
				return GetFieldValue(v, left.Name.String()).(float64) == rightValue
			case sqlparser.StrVal:
				return GetFieldValue(v, left.Name.String()).(string) == string(right.Val)
			default:
				panic("")
			}
		case sqlparser.LessThanStr:
			switch right.Type {
			case sqlparser.IntVal:
				rightValue, _ := strconv.Atoi(string(right.Val))
				return GetFieldValue(v, left.Name.String()).(int) < rightValue
			case sqlparser.FloatVal:
				rightValue, _ := strconv.ParseFloat(string(right.Val), 64)
				return GetFieldValue(v, left.Name.String()).(float64) < rightValue
			case sqlparser.StrVal:
				return GetFieldValue(v, left.Name.String()).(string) < string(right.Val)
			default:
				panic("")
			}
		case sqlparser.GreaterThanStr:
			switch right.Type {
			case sqlparser.IntVal:
				rightValue, _ := strconv.Atoi(string(right.Val))
				return GetFieldValue(v, left.Name.String()).(int) > rightValue
			case sqlparser.FloatVal:
				rightValue, _ := strconv.ParseFloat(string(right.Val), 64)
				return GetFieldValue(v, left.Name.String()).(float64) > rightValue
			case sqlparser.StrVal:
				return GetFieldValue(v, left.Name.String()).(string) > string(right.Val)
			default:
				panic("")
			}
		case sqlparser.LessEqualStr:
			switch right.Type {
			case sqlparser.IntVal:
				rightValue, _ := strconv.Atoi(string(right.Val))
				return GetFieldValue(v, left.Name.String()).(int) <= rightValue
			case sqlparser.FloatVal:
				rightValue, _ := strconv.ParseFloat(string(right.Val), 64)
				return GetFieldValue(v, left.Name.String()).(float64) <= rightValue
			case sqlparser.StrVal:
				return GetFieldValue(v, left.Name.String()).(string) <= string(right.Val)
			default:
				panic("")
			}
		case sqlparser.GreaterEqualStr:
			switch right.Type {
			case sqlparser.IntVal:
				rightValue, _ := strconv.Atoi(string(right.Val))
				return GetFieldValue(v, left.Name.String()).(int) >= rightValue
			case sqlparser.FloatVal:
				rightValue, _ := strconv.ParseFloat(string(right.Val), 64)
				return GetFieldValue(v, left.Name.String()).(float64) >= rightValue
			case sqlparser.StrVal:
				return GetFieldValue(v, left.Name.String()).(string) >= string(right.Val)
			default:
				panic("")
			}
		case sqlparser.NotEqualStr:
			switch right.Type {
			case sqlparser.IntVal:
				rightValue, _ := strconv.Atoi(string(right.Val))
				return GetFieldValue(v, left.Name.String()).(int) != rightValue
			case sqlparser.FloatVal:
				rightValue, _ := strconv.ParseFloat(string(right.Val), 64)
				return GetFieldValue(v, left.Name.String()).(float64) != rightValue
			case sqlparser.StrVal:
				return GetFieldValue(v, left.Name.String()).(string) != string(right.Val)
			default:
				panic("")
			}
		default:
			panic("unsupported operator: " + compare.Operator)
		}
	}
}

func conditionParser(expr sqlparser.Expr) func(v interface{}) bool {
	switch condition := expr.(type) {
	case *sqlparser.ComparisonExpr:
		// 条件比较的最小单元
		return parseComparison(condition)
	case *sqlparser.AndExpr:
		return func(v interface{}) bool {
			return conditionParser(condition.Left)(v) && conditionParser(condition.Right)(v)
		}
	case *sqlparser.OrExpr:
		return func(v interface{}) bool {
			return conditionParser(condition.Left)(v) || conditionParser(condition.Right)(v)
		}
	case *sqlparser.ParenExpr:
		return conditionParser(condition.Expr)
	default:
		panic("unsupported condition: " + sqlparser.String(condition))
	}
}

func TestParse(t *testing.T) {
	// sql := "select a.Symbol, b.Price from trade a, instrument b where a.Size > 100 and (b.MarkPrice > 0 or b.FairPrice > 0)"
	sql := "select Symbol, Price from trade where Price > 0.0"

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

		if conditionParser(stmt.Where.Expr)(&td) {
			t.Fatal("failed.")
		}
	default:
		t.Fatal("unsupported statement: " + sqlparser.String(stmt))
	}
}
