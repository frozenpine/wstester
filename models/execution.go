package models

import (
	"encoding/json"

	"github.com/frozenpine/ngerest"
)

// ExecutionResponse execer response structure
type ExecutionResponse struct {
	tableResponse

	Data []*ngerest.Execution `json:"data"`
}

func (exec *ExecutionResponse) String() string {
	result, _ := json.Marshal(exec)

	return string(result)
}

// Format format String output
func (exec *ExecutionResponse) Format(format string) string {
	return exec.String()
}

// GetAction get action for response
func (exec *ExecutionResponse) GetAction() string {
	return exec.Action
}

// GetData get data for reponse
func (exec *ExecutionResponse) GetData() []interface{} {
	var data []interface{}

	for _, d := range exec.Data {
		data = append(data, d)
	}

	return data
}
