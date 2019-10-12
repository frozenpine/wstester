package utils

import (
	"errors"
	"log"
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
	return reflect.Indirect(reflect.ValueOf(data)).FieldByName(property).Interface()
}

// GetFieldType get property type in struct
func GetFieldType(data interface{}, property string) *reflect.StructField {
	t := reflect.TypeOf(data)

	if typ, exist := t.FieldByName(property); exist {
		return &typ
	}

	return nil
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

// GetFieldTypes get properties type in struct
func GetFieldTypes(data interface{}, properties ...string) map[string]*reflect.StructField {
	result := make(map[string]*reflect.StructField)

	t := reflect.TypeOf(data)

	for _, prop := range properties {
		if typ, exist := t.FieldByName(prop); exist {
			result[prop] = &typ
		}
	}

	return result
}

func parseTables(stmt *sqlparser.Select) (map[string]*sqlparser.AliasedTableExpr, error) {
	tableDefine := make(map[string]*sqlparser.AliasedTableExpr)

	for _, table := range stmt.From {
		table, ok := table.(*sqlparser.AliasedTableExpr)
		if !ok {
			return nil, errors.New("invalid table type")
		}

		tableName, ok := table.Expr.(sqlparser.TableName)
		if !ok {
			return nil, errors.New("invalid table type")
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
			return nil, errors.New("Unsupported table: " + nameStr)
		}

		if !table.As.IsEmpty() {
			nameStr = table.As.String()
		}
		tableDefine[nameStr] = table
	}

	return tableDefine, nil
}

func parseColumns(tables map[string]*sqlparser.AliasedTableExpr, stmt *sqlparser.Select) (map[string][]string, error) {
	columeDefine := make(map[string][]string)

	for _, column := range stmt.SelectExprs {
		column, ok := column.(*sqlparser.AliasedExpr)
		if !ok {
			return nil, errors.New("invalid column statement")
		}

		columnName, ok := column.Expr.(*sqlparser.ColName)

		var (
			tblName string
		)

		if columnName.Qualifier.IsEmpty() {
			if len(tables) != 1 {
				log.Fatal("columns has invalid table qualifier.")
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

	return columeDefine, nil
}

func parseCondition(tables map[string]*sqlparser.AliasedTableExpr, stmt *sqlparser.Select) (map[string]func(interface{}) bool, error) {
	condition := make(map[string]func(interface{}) bool)

	// for _, condition := range stmt.Where {

	// }

	return condition, nil
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
		tableDefine, err := parseTables(stmt)
		if err != nil {
			return nil, err
		}

		parseColumns(tableDefine, stmt)
	default:
		return nil, errors.New("sql statement must be SELECT")
	}

	return filterFunc, nil
}
