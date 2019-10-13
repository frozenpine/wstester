package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	linq "github.com/ahmetb/go-linq"
	"github.com/xwb1989/sqlparser"
)

// LinqFilter filter function powered by linq
type LinqFilter func(interface{}) []map[string]interface{}

var (
	tableModels map[string]interface{}
)

// RegisterTableModel register table module for linq query
func RegisterTableModel(name string, tbl interface{}) error {
	if tableModels == nil {
		tableModels = make(map[string]interface{})
	} else if _, exist := tableModels[name]; exist {
		return fmt.Errorf("model for %s is already exist", name)
	}

	tableModels[name] = tbl

	return nil
}

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

	v := reflect.Indirect(reflect.ValueOf(data))
	t := reflect.TypeOf(v.Interface())

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

	t := reflect.TypeOf(reflect.Indirect(reflect.ValueOf(data)).Interface())

	for _, prop := range properties {
		if prop == "*" {
			for idx := 0; idx < t.NumField(); idx++ {
				typ := t.Field(idx)

				result[typ.Name] = &typ
			}

			break
		}

		if typ, exist := t.FieldByName(prop); exist {
			result[prop] = &typ
		}
	}

	return result
}

// ColumnDef column definitions generated from SQL & model
type ColumnDef struct {
	name, alias string
	define      *reflect.StructField
}

// GetName to get column's origin name defined by model
func (col *ColumnDef) GetName() string {
	return col.define.Name
}

// GetJSONName to get column's json name specified by model
func (col *ColumnDef) GetJSONName() string {
	jsn, exist := col.define.Tag.Lookup("json")

	if exist {
		return strings.Split(jsn, ",")[0]
	}

	return col.GetName()
}

// HasAlias to test wether column has alias definition
func (col *ColumnDef) HasAlias() bool {
	return col.alias != ""
}

// GetAliasName get column alias name, if no AS specified, return column's origin name
func (col *ColumnDef) GetAliasName() string {
	if col.HasAlias() {
		return col.alias
	}

	return col.GetName()
}

// GetType get colume type in model
func (col *ColumnDef) GetType() reflect.Type {
	return col.define.Type
}

// TableDef table definitions generated from SQL & model
type TableDef struct {
	name, alias string
	columns     map[string]*ColumnDef
	selected    map[string]*ColumnDef
	define      reflect.Type
}

// GetName get table's origin name defined by RegisterTableModel
func (tbl *TableDef) GetName() string {
	return tbl.name
}

// HasAlias determine wether table has alias name defined by SQL
func (tbl *TableDef) HasAlias() bool {
	return tbl.alias != ""
}

// GetAliasName get table's alias name, if no alias name defined, return origin name.
func (tbl *TableDef) GetAliasName() string {
	if tbl.HasAlias() {
		return tbl.alias
	}

	return tbl.GetName()
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

func parseTable(stmt *sqlparser.Select) (*TableDef, error) {
	tableDefine := TableDef{
		columns: make(map[string]*ColumnDef),
	}

	errTableType := errors.New("invalid table type")

	if len(stmt.From) > 1 {
		return nil, errors.New("only one table in select")
	}

	table, ok := stmt.From[0].(*sqlparser.AliasedTableExpr)
	if !ok {
		return nil, errTableType
	}

	tableName, ok := table.Expr.(sqlparser.TableName)
	if !ok {
		return nil, errTableType
	}

	model, exist := tableModels[tableName.Name.String()]
	if !exist {
		return nil, errTableType
	}

	tableDefine.name = tableName.Name.String()
	if !table.As.IsEmpty() {
		tableDefine.alias = table.As.String()
	}
	tableDefine.define = reflect.TypeOf(reflect.Indirect(reflect.ValueOf(model)).Interface())

	for colName, col := range GetFieldTypes(reflect.Indirect(reflect.ValueOf(model)).Interface(), "*") {
		tableDefine.columns[colName] = &ColumnDef{
			name:   colName,
			define: col,
		}
	}

	columns, err := parseColumns(&tableDefine, stmt)
	if err != nil {
		return nil, err
	}
	tableDefine.selected = columns

	return &tableDefine, nil
}

func parseColumns(tbl *TableDef, stmt *sqlparser.Select) (map[string]*ColumnDef, error) {
	var (
		selectedColumns    = make(map[string]*ColumnDef)
		errColumnName      = errors.New("invalid column name")
		errColumnStatement = errors.New("invalid column statement")
	)

	for _, col := range stmt.SelectExprs {
		switch col := col.(type) {
		case *sqlparser.AliasedExpr:
			colName, _ := col.Expr.(*sqlparser.ColName)

			colDef, exist := tbl.columns[colName.Name.String()]
			if !exist {
				return nil, errColumnName
			}
			if !col.As.IsEmpty() {
				colDef.alias = col.As.String()
			}

			selectedColumns[colDef.GetName()] = colDef
		case *sqlparser.StarExpr:
			for _, column := range tbl.columns {
				selectedColumns[column.GetName()] = column
			}
		default:
			return nil, errColumnStatement
		}
	}

	return selectedColumns, nil
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

func parseCondition(expr sqlparser.Expr) (func(v interface{}) bool, error) {
	switch condition := expr.(type) {
	case *sqlparser.ComparisonExpr:
		// 条件比较的最小单元
		return parseComparison(condition)
	case *sqlparser.AndExpr:
		leftFn, err := parseCondition(condition.Left)
		if err != nil {
			return nil, err
		}

		rightFn, err := parseCondition(condition.Right)
		if err != nil {
			return nil, err
		}
		return func(v interface{}) bool {
			return leftFn(v) && rightFn(v)
		}, nil
	case *sqlparser.OrExpr:
		leftFn, err := parseCondition(condition.Left)
		if err != nil {
			return nil, err
		}

		rightFn, err := parseCondition(condition.Right)
		if err != nil {
			return nil, err
		}

		return func(v interface{}) bool {
			return leftFn(v) || rightFn(v)
		}, nil
	case *sqlparser.ParenExpr:
		return parseCondition(condition.Expr)
	default:
		return nil, errors.New("unsupported condition: " + sqlparser.String(condition))
	}
}

// ParseSQL parse table, column & conditions from SQL
func ParseSQL(sql string) (map[string]LinqFilter, error) {
	stmt, err := sqlparser.Parse(sql)

	if err != nil {
		return nil, err
	}

	filterFunc := make(map[string]LinqFilter)

	for _, sel := range parseUnion(stmt) {
		tblDefine, err := parseTable(sel)

		if err != nil {
			return nil, err
		}

		conditionFn := func(interface{}) bool { return true }

		if sel.Where != nil {
			conditionFn, err = parseCondition(sel.Where.Expr)
			if err != nil {
				return nil, err
			}
		}

		filterFunc[tblDefine.name] = func(datas interface{}) []map[string]interface{} {
			var results []map[string]interface{}

			linq.From(datas).Where(func(v interface{}) bool {
				return conditionFn(v)
			}).Select(func(v interface{}) interface{} {
				result := make(map[string]interface{})

				for _, col := range tblDefine.selected {
					key := col.GetJSONName()
					if col.HasAlias() {
						key = col.GetAliasName()
					}
					result[key] = GetFieldValue(v, col.GetName())
				}

				return result
			}).ToSlice(&results)

			return results
		}
	}

	return filterFunc, nil
}
