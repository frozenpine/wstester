package ngerest

// ModelError error
type ModelError struct {
	Error *ErrorError `json:"error"`
}
