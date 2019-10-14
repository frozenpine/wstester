package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// MarginResponse marginer response structure
type MarginResponse struct {
	tableResponse

	Data []*ngerest.Margin `json:"data"`
}

func (margin *MarginResponse) String() string {
	result, _ := json.Marshal(margin)

	return string(result)
}

// Format format String output
func (margin *MarginResponse) Format(format string) string {
	return margin.String()
}

// GetAction get action for response
func (margin *MarginResponse) GetAction() string {
	return margin.Action
}

// GetData get data for reponse
func (margin *MarginResponse) GetData() []interface{} {
	var data []interface{}

	for _, d := range margin.Data {
		data = append(data, d)
	}

	return data
}
