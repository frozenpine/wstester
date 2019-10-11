package utils

import (
	"errors"
	"reflect"
	"strings"

	"github.com/frozenpine/wstester/models"
	"github.com/xwb1989/sqlparser"
)

// TODO: SQL to LINQ

var (
	tableMapper = make(map[string]interface{})
)

// GetFieldValue get property value in struct
func GetFieldValue(data interface{}, property string) interface{} {
	return reflect.ValueOf(data).FieldByName(property).Interface()
}

// GetFields get properties value in struct
func GetFields(data interface{}, properties ...string) map[string]interface{} {
	result := make(map[string]interface{})

	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	for _, prop := range properties {
		field := v.FieldByName(prop)
		fieldType, _ := t.FieldByName(prop)

		if tag := fieldType.Tag.Get("json"); tag != "" {
			prop = strings.Split(tag, ",")[0]
		}

		result[prop] = field.Interface()
	}

	return result
}

func parseTables(stmt *sqlparser.Select) map[string]*sqlparser.AliasedTableExpr {
	tableDefine := make(map[string]*sqlparser.AliasedTableExpr)

	for _, table := range stmt.From {
		table, ok := table.(*sqlparser.AliasedTableExpr)
		if !ok {
			panic("invalid table type")
		}

		tableName, ok := table.Expr.(sqlparser.TableName)
		if !ok {
			panic("invalid table type")
		}

		nameStr := tableName.Name.String()

		switch nameStr {
		case "trade":
			tableMapper[nameStr] = new(models.TradeResponse)
		case "instrument":
			tableMapper[nameStr] = new(models.InstrumentResponse)
		case "orderBookL2":
			tableMapper[nameStr] = new(models.MBLResponse)
		default:
			panic("unsupported table: " + nameStr)
		}

		if !table.As.IsEmpty() {
			nameStr = table.As.String()
		}
		tableDefine[nameStr] = table
	}

	return tableDefine
}

func parseColumns(tables map[string]*sqlparser.AliasedTableExpr, stmt *sqlparser.Select) map[string][]string {
	columeDefine := make(map[string][]string)

	for _, column := range stmt.SelectExprs {
		column, ok := column.(*sqlparser.AliasedExpr)
		if !ok {
			panic("invalid column statement.")
		}

		columnName, ok := column.Expr.(*sqlparser.ColName)

		var (
			tblName string
		)

		if columnName.Qualifier.IsEmpty() {
			if len(tables) != 1 {
				panic("columns has invalid table qualifier.")
			}

			for tblName = range tables {
				break
			}
		} else {
			tblName = columnName.Qualifier.Name.String()
		}

		if _, exist := tables[tblName]; !exist {
			panic("columns has no table.")
		} else {
			columeDefine[tblName] = append(columeDefine[tblName], columnName.Name.String())
		}
	}

	return columeDefine
}

func parseCondition(tables map[string]*sqlparser.AliasedTableExpr, stmt *sqlparser.Select) map[string]func(interface{}) bool {
	condition := make(map[string]func(interface{}) bool)

	return condition
}

// ParseSQL parse table, column & conditions from SQL
func ParseSQL(sql string) (map[string]func(models.Response) map[string]interface{}, error) {
	stmt, err := sqlparser.Parse(sql)

	if err != nil {
		return nil, err
	}

	filterFunc := make(map[string]func(models.Response) map[string]interface{})

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		tableDefine := parseTables(stmt)

		parseColumns(tableDefine, stmt)

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

		// for _, condition := range stmt.Where {

		// }
	default:
		return nil, errors.New("sql statement must be SELECT")
	}

	return filterFunc, nil
}
