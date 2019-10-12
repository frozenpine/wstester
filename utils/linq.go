package utils

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/frozenpine/wstester/models"
	"github.com/xwb1989/sqlparser"
)

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

	errTableType := errors.New("invalid table type")

	for _, table := range stmt.From {
		table, ok := table.(*sqlparser.AliasedTableExpr)
		if !ok {
			return nil, errTableType
		}

		tableName, ok := table.Expr.(sqlparser.TableName)
		if !ok {
			return nil, errTableType
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
		switch column := column.(type) {
		case *sqlparser.AliasedExpr:
			columnName, _ := column.Expr.(*sqlparser.ColName)

			var (
				tblName string
			)

			if columnName.Qualifier.IsEmpty() {
				if len(tables) != 1 {
					return nil, errors.New("columns has invalid table qualifier")
				}

				for tblName = range tables {
					break
				}
			} else {
				tblName = columnName.Qualifier.Name.String()
			}

			if _, exist := tables[tblName]; !exist {
				return nil, errors.New("columns has no table")
			}

			columeDefine[tblName] = append(columeDefine[tblName], columnName.Name.String())
		case *sqlparser.StarExpr:

		default:
			return nil, errors.New("invalid column statement")
		}
	}

	return columeDefine, nil
}

func parseInt(left interface{}, right *sqlparser.SQLVal) (int, int) {
	leftValue := left.(int)
	rightValue, _ := strconv.Atoi(string(right.Val))

	return leftValue, rightValue
}

func parseFloat(left interface{}, right *sqlparser.SQLVal) (float64, float64) {
	leftValue := left.(float64)
	rightValue, _ := strconv.ParseFloat(string(right.Val), 64)

	return leftValue, rightValue
}

func parseStr(left interface{}, right *sqlparser.SQLVal) (string, string) {
	leftValue := left.(string)
	rightValue := string(right.Val)

	return leftValue, rightValue
}

func parseComparison(compare *sqlparser.ComparisonExpr) (func(interface{}) bool, error) {
	left, ok := compare.Left.(*sqlparser.ColName)
	if !ok {
		return nil, errors.New("left side must be a property of struct")
	}

	right, ok := compare.Right.(*sqlparser.SQLVal)
	if !ok {
		return nil, errors.New("right side must be a literal value")
	}

	return func(v interface{}) bool {
		errOperator := errors.New("unsupported operator: " + compare.Operator)
		errValueType := errors.New("invalid value type")

		switch compare.Operator {
		case sqlparser.EqualStr:
			switch right.Type {
			case sqlparser.IntVal:
				leftValue, rightValue := parseInt(GetFieldValue(v, left.Name.String()), right)
				return leftValue == rightValue
			case sqlparser.FloatVal:
				leftValue, rightValue := parseFloat(GetFieldValue(v, left.Name.String()), right)
				return leftValue == rightValue
			case sqlparser.StrVal:
				leftValue, rightValue := parseStr(GetFieldValue(v, left.Name.String()), right)
				return leftValue == rightValue
			}
		case sqlparser.LessThanStr:
			switch right.Type {
			case sqlparser.IntVal:
				leftValue, rightValue := parseInt(GetFieldValue(v, left.Name.String()), right)
				return leftValue < rightValue
			case sqlparser.FloatVal:
				leftValue, rightValue := parseFloat(GetFieldValue(v, left.Name.String()), right)
				return leftValue < rightValue
			case sqlparser.StrVal:
				leftValue, rightValue := parseStr(GetFieldValue(v, left.Name.String()), right)
				return leftValue < rightValue
			}
		case sqlparser.GreaterThanStr:
			switch right.Type {
			case sqlparser.IntVal:
				leftValue, rightValue := parseInt(GetFieldValue(v, left.Name.String()), right)
				return leftValue > rightValue
			case sqlparser.FloatVal:
				leftValue, rightValue := parseFloat(GetFieldValue(v, left.Name.String()), right)
				return leftValue > rightValue
			case sqlparser.StrVal:
				leftValue, rightValue := parseStr(GetFieldValue(v, left.Name.String()), right)
				return leftValue > rightValue
			}
		case sqlparser.LessEqualStr:
			switch right.Type {
			case sqlparser.IntVal:
				leftValue, rightValue := parseInt(GetFieldValue(v, left.Name.String()), right)
				return leftValue <= rightValue
			case sqlparser.FloatVal:
				leftValue, rightValue := parseFloat(GetFieldValue(v, left.Name.String()), right)
				return leftValue <= rightValue
			case sqlparser.StrVal:
				leftValue, rightValue := parseStr(GetFieldValue(v, left.Name.String()), right)
				return leftValue <= rightValue
			}
		case sqlparser.GreaterEqualStr:
			switch right.Type {
			case sqlparser.IntVal:
				leftValue, rightValue := parseInt(GetFieldValue(v, left.Name.String()), right)
				return leftValue >= rightValue
			case sqlparser.FloatVal:
				leftValue, rightValue := parseFloat(GetFieldValue(v, left.Name.String()), right)
				return leftValue >= rightValue
			case sqlparser.StrVal:
				leftValue, rightValue := parseStr(GetFieldValue(v, left.Name.String()), right)
				return leftValue >= rightValue
			}
		case sqlparser.NotEqualStr:
			switch right.Type {
			case sqlparser.IntVal:
				leftValue, rightValue := parseInt(GetFieldValue(v, left.Name.String()), right)
				return leftValue != rightValue
			case sqlparser.FloatVal:
				leftValue, rightValue := parseFloat(GetFieldValue(v, left.Name.String()), right)
				return leftValue != rightValue
			case sqlparser.StrVal:
				leftValue, rightValue := parseStr(GetFieldValue(v, left.Name.String()), right)
				return leftValue != rightValue
			}
		default:
			panic(errOperator)
		}

		panic(errValueType)
	}, nil
}

func conditionParser(expr sqlparser.Expr) (func(v interface{}) bool, error) {
	switch condition := expr.(type) {
	case *sqlparser.ComparisonExpr:
		// 条件比较的最小单元
		return parseComparison(condition)
	case *sqlparser.AndExpr:
		leftFn, err := conditionParser(condition.Left)
		if err != nil {
			return nil, err
		}

		rightFn, err := conditionParser(condition.Right)
		if err != nil {
			return nil, err
		}
		return func(v interface{}) bool {
			return leftFn(v) && rightFn(v)
		}, nil
	case *sqlparser.OrExpr:
		leftFn, err := conditionParser(condition.Left)
		if err != nil {
			return nil, err
		}

		rightFn, err := conditionParser(condition.Right)
		if err != nil {
			return nil, err
		}

		return func(v interface{}) bool {
			return leftFn(v) || rightFn(v)
		}, nil
	case *sqlparser.ParenExpr:
		return conditionParser(condition.Expr)
	default:
		return nil, errors.New("unsupported condition: " + sqlparser.String(condition))
	}
}

func parseUnion(stmt sqlparser.Statement) []*sqlparser.Select {
	var result []*sqlparser.Select

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		result = append(result, stmt)
	case *sqlparser.Union:
		if rightSelect := parseUnion(stmt.Right); rightSelect != nil {
			result = append(result, rightSelect...)
		}

		if leftSelect := parseUnion(stmt.Left); leftSelect != nil {
			result = append(result, leftSelect...)
		}
	}

	return result
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
