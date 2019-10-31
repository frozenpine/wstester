package models

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/frozenpine/ngerest"
)

// InstrumentResponse instrument response structure
type InstrumentResponse struct {
	tableResponse

	Data []*ngerest.Instrument `json:"data"`
}

// NewInstrumentPartial create an new instrument partial response
func NewInstrumentPartial() *InstrumentResponse {
	partial := InstrumentResponse{}

	partial.Table = "instrument"
	partial.Action = PartialAction
	partial.Keys = []string{}
	partial.Types = make(map[string]string)
	partial.ForeignKeys = make(map[string]string)
	partial.Attributes = make(map[string]string)
	partial.Filter = make(map[string]string)

	return &partial
}

// String get structure's string format
func (ins *InstrumentResponse) String() string {
	result, _ := json.Marshal(ins)

	return string(result)
}

// Format format String output
func (ins *InstrumentResponse) Format(format string) string {
	if format == "" {
		return ins.String()
	}

	tpl, err := template.New("insRsp").Parse(format)
	if err != nil {
		panic(err)
	}

	buf := bytes.Buffer{}

	if err := tpl.Execute(&buf, ins); err != nil {
		panic(err)
	}

	return buf.String()
}

// GetAction get action for response
func (ins *InstrumentResponse) GetAction() string {
	return ins.Action
}

// GetData get data for reponse
func (ins *InstrumentResponse) GetData() []interface{} {
	var data []interface{}

	for _, d := range ins.Data {
		data = append(data, d)
	}

	return data
}
