package utils

import (
	"errors"
	"fmt"
	"log"
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

	operatorMapper = map[string][3]func(interface{}, string, *sqlparser.SQLVal) bool{
		sqlparser.EqualStr: [3]func(interface{}, string, *sqlparser.SQLVal) bool{
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseStr(GetFieldValue(l, lName), r)
				return lValue == rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseInt(GetFieldValue(l, lName), r)
				return lValue == rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseFloat(GetFieldValue(l, lName), r)
				return lValue == rValue
			},
		},
		sqlparser.LessThanStr: [3]func(interface{}, string, *sqlparser.SQLVal) bool{
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseStr(GetFieldValue(l, lName), r)
				return lValue < rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseInt(GetFieldValue(l, lName), r)
				return lValue < rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseFloat(GetFieldValue(l, lName), r)
				return lValue < rValue
			},
		},
		sqlparser.GreaterThanStr: [3]func(interface{}, string, *sqlparser.SQLVal) bool{
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseStr(GetFieldValue(l, lName), r)
				return lValue > rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseInt(GetFieldValue(l, lName), r)
				return lValue > rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseFloat(GetFieldValue(l, lName), r)
				return lValue > rValue
			},
		},
		sqlparser.LessEqualStr: [3]func(interface{}, string, *sqlparser.SQLVal) bool{
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseStr(GetFieldValue(l, lName), r)
				return lValue <= rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseInt(GetFieldValue(l, lName), r)
				return lValue <= rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseFloat(GetFieldValue(l, lName), r)
				return lValue <= rValue
			},
		},
		sqlparser.GreaterEqualStr: [3]func(interface{}, string, *sqlparser.SQLVal) bool{
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseStr(GetFieldValue(l, lName), r)
				return lValue >= rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseInt(GetFieldValue(l, lName), r)
				return lValue >= rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseFloat(GetFieldValue(l, lName), r)
				return lValue >= rValue
			},
		},
		sqlparser.NotEqualStr: [3]func(interface{}, string, *sqlparser.SQLVal) bool{
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseStr(GetFieldValue(l, lName), r)
				return lValue != rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseInt(GetFieldValue(l, lName), r)
				return lValue != rValue
			},
			func(l interface{}, lName string, r *sqlparser.SQLVal) bool {
				lValue, rValue := parseFloat(GetFieldValue(l, lName), r)
				return lValue != rValue
			},
		},
	}
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
	if property == "" {
		log.Panicln("property name can not be null")
	}

	return reflect.Indirect(reflect.ValueOf(data)).FieldByName(property).Interface()
}

// GetFieldType get property type in struct
func GetFieldType(data interface{}, property string) *reflect.StructField {
	t := reflect.TypeOf(reflect.Indirect(reflect.ValueOf(data)).Interface())

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
	where       func(interface{}) bool
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

// GetFilter get a filter function for data slice
func (tbl *TableDef) GetFilter() LinqFilter {
	return func(datas interface{}) []map[string]interface{} {
		var results []map[string]interface{}

		query := linq.From(datas)

		if tbl.where != nil {
			query = query.Where(func(v interface{}) bool {
				return tbl.where(v)
			})
		}

		query.Select(func(v interface{}) interface{} {
			if reflect.Indirect(reflect.ValueOf(v)).Interface() == nil {
				return v
			}

			result := make(map[string]interface{})

			for _, col := range tbl.selected {
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

	if len(stmt.From) > 1 {
		return nil, errors.New("only support one table in FROM statement")
	}

	table, ok := stmt.From[0].(*sqlparser.AliasedTableExpr)
	if !ok {
		return nil, errors.New("invalid table type: " + sqlparser.String(stmt.From[0]))
	}

	tableName, ok := table.Expr.(sqlparser.TableName)
	if !ok {
		return nil, errors.New("can not convert SQLNode to TableName")
	}

	model, exist := tableModels[tableName.Name.String()]
	if !exist {
		return nil, errors.New("table model not defined: " + tableName.Name.String())
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
		selectedColumns = make(map[string]*ColumnDef)
	)

	for _, col := range stmt.SelectExprs {
		switch col := col.(type) {
		case *sqlparser.AliasedExpr:
			colName, _ := col.Expr.(*sqlparser.ColName)

			key := strings.Title(colName.Name.String())

			colDef, exist := tbl.columns[key]
			if !exist {
				return nil, fmt.Errorf("column[%s] definition not exists", colName.Name.String())
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
			return nil, fmt.Errorf("unsupported column statement: %s", sqlparser.String(col))
		}
	}

	return selectedColumns, nil
}

func parseInt(left interface{}, right *sqlparser.SQLVal) (int64, int64) {
	leftValue := reflect.ValueOf(left).Convert(reflect.TypeOf(int64(1))).Interface().(int64)
	rightValue, _ := strconv.ParseInt(string(right.Val), 10, 64)

	return leftValue, rightValue
}

func parseFloat(left interface{}, right *sqlparser.SQLVal) (float64, float64) {
	leftValue := reflect.ValueOf(left).Convert(reflect.TypeOf(float64(1))).Interface().(float64)

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

	leftName := strings.Title(left.Name.String())
	if leftName == "Id" {
		leftName = "ID"
	}

	right, ok := compare.Right.(*sqlparser.SQLVal)
	if !ok {
		return nil, errors.New("right side must be a literal value")
	}

	return func(v interface{}) bool {
		fnList, exist := operatorMapper[compare.Operator]
		if !exist {
			panic("unsupported operator: " + compare.Operator)
		}

		if len(fnList) <= int(right.Type) {
			panic("invalid value type")
		}

		return fnList[right.Type](v, leftName, right)
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
func ParseSQL(sql string) (map[string]*TableDef, error) {
	tables := make(map[string]*TableDef)

	if sql == "" {
		return tables, nil
	}

	stmt, err := sqlparser.Parse(sql)

	if err != nil {
		return nil, err
	}

	for _, sel := range parseUnion(stmt) {
		tblDefine, err := parseTable(sel)

		if err != nil {
			return nil, err
		}

		if sel.Where != nil {
			conditionFn, err := parseCondition(sel.Where.Expr)
			if err != nil {
				return nil, err
			}

			tblDefine.where = conditionFn
		}

		tables[tblDefine.GetName()] = tblDefine
	}

	return tables, nil
}
