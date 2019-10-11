package utils

import (
	"errors"

	"github.com/frozenpine/wstester/models"
	"github.com/xwb1989/sqlparser"
)

// TODO: SQL to LINQ

var (
	tableMapper = make(map[string]models.Response)
)

// ParseSQL parse table, column & conditions from SQL
func ParseSQL(sql string) (map[string]func(models.Response) map[string]interface{}, error) {
	stmt, err := sqlparser.Parse(sql)

	if err != nil {
		return nil, err
	}

	filterFunc := make(map[string]func(models.Response) map[string]interface{})

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		// tableColumn := make(map[string][]string)

		// for _, column := range stmt.SelectExprs {
		// 	column := column.(*sqlparser.AliasedExpr)
		// 	columnName := column.Expr.(*sqlparser.ColName)
		// 	columnAs := column.As

		// 	if columnName.Qualifier.IsEmpty() {
		// 		tableColumn[columnName]
		// 	}

		// 	switch columnName.Qualifier.IsEmpty() {
		// 	case condition:

		// 	}
		// }

		for _, table := range stmt.From {
			table, ok := table.(*sqlparser.AliasedTableExpr)
			if !ok {
				return nil, errors.New("invalid table type")
			}

			tableName, ok := table.Expr.(sqlparser.TableName)
			if !ok {
				return nil, errors.New("invalid table type")
			}
			// tableAs := table.As

			switch tableName.Name.String() {
			case "trade":
				tableMapper["trade"] = new(models.TradeResponse)
			case "instrument":
				tableMapper["instrument"] = new(models.InstrumentResponse)
			case "orderBookL2":
				tableMapper["orderBookL2"] = new(models.MBLResponse)
			default:
				return nil, errors.New("unsupported table: " + tableName.Name.String())
			}
		}

		// for _, condition := range stmt.Where {

		// }
	default:
		return nil, errors.New("sql statement must be SELECT")
	}

	return filterFunc, nil
}
