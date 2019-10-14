package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// OrderResponse order response structure
type OrderResponse struct {
	tableResponse

	Data []*ngerest.Order `json:"data"`
}

func (ord *OrderResponse) String() string {
	result, _ := json.Marshal(ord)

	return string(result)
}

// Format format String output
func (ord *OrderResponse) Format(format string) string {
	return ord.String()
}

// GetAction get action for response
func (ord *OrderResponse) GetAction() string {
	return ord.Action
}

// GetData get data for reponse
func (ord *OrderResponse) GetData() []interface{} {
	var data []interface{}

	for _, d := range ord.Data {
		data = append(data, d)
	}

	return data
}
