package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// PositionResponse poser response structure
type PositionResponse struct {
	tableResponse

	Data []*ngerest.Position `json:"data"`
}

func (pos *PositionResponse) String() string {
	result, _ := json.Marshal(pos)

	return string(result)
}

// Format format String output
func (pos *PositionResponse) Format(format string) string {
	return pos.String()
}

// GetAction get action for response
func (pos *PositionResponse) GetAction() string {
	return pos.Action
}

// GetData get data for reponse
func (pos *PositionResponse) GetData() []interface{} {
	var data []interface{}

	for _, d := range pos.Data {
		data = append(data, d)
	}

	return data
}
