package ngerest

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/antihax/optional"
)

// Linger please
var (
	_ context.Context
)

// OrderAPIService order api service
type OrderAPIService service

// OrderAmendOpts optinal args for amend order
type OrderAmendOpts struct {
	OrderID         optional.String
	OrigClOrdID     optional.String
	ClOrdID         optional.String
	SimpleOrderQty  optional.Float64
	OrderQty        optional.Float32
	SimpleLeavesQty optional.Float64
	LeavesQty       optional.Float32
	Price           optional.Float64
	StopPx          optional.Float64
	PegOffsetValue  optional.Float64
	Text            optional.String
}

// OrderAmend amend the quantity or price of an open order.
func (a *OrderAPIService) OrderAmend(ctx context.Context, localVarOptionals *OrderAmendOpts) (Order, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("Put")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue Order
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/order"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/xml", "text/xml", "application/javascript", "text/javascript"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.OrderID.IsSet() {
		localVarFormParams.Add("orderID", parameterToString(localVarOptionals.OrderID.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.OrigClOrdID.IsSet() {
		localVarFormParams.Add("origClOrdID", parameterToString(localVarOptionals.OrigClOrdID.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.ClOrdID.IsSet() {
		localVarFormParams.Add("clOrdID", parameterToString(localVarOptionals.ClOrdID.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.SimpleOrderQty.IsSet() {
		localVarFormParams.Add("simpleOrderQty", parameterToString(localVarOptionals.SimpleOrderQty.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.OrderQty.IsSet() {
		localVarFormParams.Add("orderQty", parameterToString(localVarOptionals.OrderQty.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.SimpleLeavesQty.IsSet() {
		localVarFormParams.Add("simpleLeavesQty", parameterToString(localVarOptionals.SimpleLeavesQty.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.LeavesQty.IsSet() {
		localVarFormParams.Add("leavesQty", parameterToString(localVarOptionals.LeavesQty.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Price.IsSet() {
		localVarFormParams.Add("price", parameterToString(localVarOptionals.Price.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.StopPx.IsSet() {
		localVarFormParams.Add("stopPx", parameterToString(localVarOptionals.StopPx.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.PegOffsetValue.IsSet() {
		localVarFormParams.Add("pegOffsetValue", parameterToString(localVarOptionals.PegOffsetValue.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Text.IsSet() {
		localVarFormParams.Add("text", parameterToString(localVarOptionals.Text.Value(), ""))
	}
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
			var key string
			if auth.Prefix != "" {
				key = auth.Prefix + " " + auth.Key
			} else {
				key = auth.Key
			}
			localVarHeaderParams["api-key"] = key

		}
	}
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-nonce"] = key

	// 	}
	// }
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-signature"] = key

	// 	}
	// }
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHTTPResponse, err
		}
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}

		if localVarHTTPResponse.StatusCode == 200 {
			var v Order
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 400 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 401 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 404 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

// OrderAmendBulkOpts optinal args for amend bulk orders
type OrderAmendBulkOpts struct {
	Orders optional.String
}

// OrderAmendBulk amend multiple orders for the same symbol.
func (a *OrderAPIService) OrderAmendBulk(ctx context.Context, localVarOptionals *OrderAmendBulkOpts) ([]Order, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("Put")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue []Order
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/order/bulk"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/xml", "text/xml", "application/javascript", "text/javascript"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.Orders.IsSet() {
		localVarFormParams.Add("orders", parameterToString(localVarOptionals.Orders.Value(), ""))
	}
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
			var key string
			if auth.Prefix != "" {
				key = auth.Prefix + " " + auth.Key
			} else {
				key = auth.Key
			}
			localVarHeaderParams["api-key"] = key

		}
	}
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-nonce"] = key

	// 	}
	// }
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-signature"] = key

	// 	}
	// }
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHTTPResponse, err
		}
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}

		if localVarHTTPResponse.StatusCode == 200 {
			var v []Order
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 400 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 401 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 404 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

// OrderCancelOpts optinal args for cancel orders
type OrderCancelOpts struct {
	OrderID optional.String
	ClOrdID optional.String
	Text    optional.String
}

// OrderCancel cancel order(s). Send multiple order IDs to cancel in bulk.
func (a *OrderAPIService) OrderCancel(ctx context.Context, localVarOptionals *OrderCancelOpts) ([]Order, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("Delete")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue []Order
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/order"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/xml", "text/xml", "application/javascript", "text/javascript"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.OrderID.IsSet() {
		localVarFormParams.Add("orderID", parameterToString(localVarOptionals.OrderID.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.ClOrdID.IsSet() {
		localVarFormParams.Add("clOrdID", parameterToString(localVarOptionals.ClOrdID.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Text.IsSet() {
		localVarFormParams.Add("text", parameterToString(localVarOptionals.Text.Value(), ""))
	}
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
			var key string
			if auth.Prefix != "" {
				key = auth.Prefix + " " + auth.Key
			} else {
				key = auth.Key
			}
			localVarHeaderParams["api-key"] = key

		}
	}
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-nonce"] = key

	// 	}
	// }
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-signature"] = key

	// 	}
	// }
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHTTPResponse, err
		}
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}

		if localVarHTTPResponse.StatusCode == 200 {
			var v []Order
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 400 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 401 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 404 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

// OrderCancelAllOpts optinal args for cancel all orders
type OrderCancelAllOpts struct {
	Symbol optional.String
	Filter optional.String
	Text   optional.String
}

// OrderCancelAll cancels all of your orders.
func (a *OrderAPIService) OrderCancelAll(ctx context.Context, localVarOptionals *OrderCancelAllOpts) ([]Order, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("Delete")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue []Order
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/order/all"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/xml", "text/xml", "application/javascript", "text/javascript"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.Symbol.IsSet() {
		localVarFormParams.Add("symbol", parameterToString(localVarOptionals.Symbol.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Filter.IsSet() {
		localVarFormParams.Add("filter", parameterToString(localVarOptionals.Filter.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Text.IsSet() {
		localVarFormParams.Add("text", parameterToString(localVarOptionals.Text.Value(), ""))
	}
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
			var key string
			if auth.Prefix != "" {
				key = auth.Prefix + " " + auth.Key
			} else {
				key = auth.Key
			}
			localVarHeaderParams["api-key"] = key

		}
	}
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-nonce"] = key

	// 	}
	// }
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-signature"] = key

	// 	}
	// }
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHTTPResponse, err
		}
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}

		if localVarHTTPResponse.StatusCode == 200 {
			var v []Order
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 400 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 401 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 404 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

// OrderCancelAllAfter automatically cancel all your orders after a specified timeout.
func (a *OrderAPIService) OrderCancelAllAfter(ctx context.Context, timeout float64) (interface{}, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("Post")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue interface{}
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/order/cancelAllAfter"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/xml", "text/xml", "application/javascript", "text/javascript"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	localVarFormParams.Add("timeout", parameterToString(timeout, ""))
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
			var key string
			if auth.Prefix != "" {
				key = auth.Prefix + " " + auth.Key
			} else {
				key = auth.Key
			}
			localVarHeaderParams["api-key"] = key

		}
	}
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-nonce"] = key

	// 	}
	// }
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-signature"] = key

	// 	}
	// }
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHTTPResponse, err
		}
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}

		if localVarHTTPResponse.StatusCode == 200 {
			var v interface{}
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 400 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 401 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 404 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

// OrderClosePositionOpts optinal args for close a position
type OrderClosePositionOpts struct {
	Price optional.Float64
}

// OrderClosePosition close a position. [Deprecated, use POST /order with execInst: Close]
func (a *OrderAPIService) OrderClosePosition(ctx context.Context, symbol string, localVarOptionals *OrderClosePositionOpts) (Order, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("Post")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue Order
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/order/closePosition"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/xml", "text/xml", "application/javascript", "text/javascript"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	localVarFormParams.Add("symbol", parameterToString(symbol, ""))
	if localVarOptionals != nil && localVarOptionals.Price.IsSet() {
		localVarFormParams.Add("price", parameterToString(localVarOptionals.Price.Value(), ""))
	}
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
			var key string
			if auth.Prefix != "" {
				key = auth.Prefix + " " + auth.Key
			} else {
				key = auth.Key
			}
			localVarHeaderParams["api-key"] = key

		}
	}
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-nonce"] = key

	// 	}
	// }
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-signature"] = key

	// 	}
	// }
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHTTPResponse, err
		}
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}

		if localVarHTTPResponse.StatusCode == 200 {
			var v Order
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 400 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 401 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 404 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

// OrderGetOrdersOpts optinal args for get orders
type OrderGetOrdersOpts struct {
	Symbol    optional.String
	Filter    optional.String
	Columns   optional.String
	Count     optional.Float32
	Start     optional.Float32
	Reverse   optional.Bool
	StartTime optional.Time
	EndTime   optional.Time
}

// OrderGetOrders get your orders.
func (a *OrderAPIService) OrderGetOrders(ctx context.Context, localVarOptionals *OrderGetOrdersOpts) ([]Order, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue []Order
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/order"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if localVarOptionals != nil && localVarOptionals.Symbol.IsSet() {
		localVarQueryParams.Add("symbol", parameterToString(localVarOptionals.Symbol.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Filter.IsSet() {
		localVarQueryParams.Add("filter", parameterToString(localVarOptionals.Filter.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Columns.IsSet() {
		localVarQueryParams.Add("columns", parameterToString(localVarOptionals.Columns.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Count.IsSet() {
		localVarQueryParams.Add("count", parameterToString(localVarOptionals.Count.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Start.IsSet() {
		localVarQueryParams.Add("start", parameterToString(localVarOptionals.Start.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Reverse.IsSet() {
		localVarQueryParams.Add("reverse", parameterToString(localVarOptionals.Reverse.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.StartTime.IsSet() {
		localVarQueryParams.Add("startTime", parameterToString(localVarOptionals.StartTime.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.EndTime.IsSet() {
		localVarQueryParams.Add("endTime", parameterToString(localVarOptionals.EndTime.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/xml", "text/xml", "application/javascript", "text/javascript"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
			var key string
			if auth.Prefix != "" {
				key = auth.Prefix + " " + auth.Key
			} else {
				key = auth.Key
			}
			localVarHeaderParams["api-key"] = key

		}
	}
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-nonce"] = key

	// 	}
	// }
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-signature"] = key

	// 	}
	// }
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHTTPResponse, err
		}
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}

		if localVarHTTPResponse.StatusCode == 200 {
			var v []Order
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 400 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 401 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 404 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

// OrderNewOpts optinal args for create a new order
type OrderNewOpts struct {
	Side            optional.String
	SimpleOrderQty  optional.Float64
	OrderQty        optional.Float32
	Price           optional.Float64
	DisplayQty      optional.Float32
	StopPx          optional.Float64
	ClOrdID         optional.String
	ClOrdLinkID     optional.String
	PegOffsetValue  optional.Float64
	PegPriceType    optional.String
	OrdType         optional.String
	TimeInForce     optional.String
	ExecInst        optional.String
	ContingencyType optional.String
	Text            optional.String
}

// OrderNew Create a new order.
func (a *OrderAPIService) OrderNew(ctx context.Context, symbol string, localVarOptionals *OrderNewOpts) (Order, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("Post")
		localVarPostBody    map[string]interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue Order
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/order"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/xml", "text/xml", "application/javascript", "text/javascript"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}

	localVarPostBody = make(map[string]interface{})
	localVarPostBody["symbol"] = parameterToString(symbol, "")

	if localVarOptionals != nil && localVarOptionals.Side.IsSet() {
		localVarPostBody["side"] = localVarOptionals.Side.Value()
	}

	if localVarOptionals != nil && localVarOptionals.SimpleOrderQty.IsSet() {
		localVarPostBody["simpleOrderQty"] = localVarOptionals.SimpleOrderQty.Value()
	}

	if localVarOptionals != nil && localVarOptionals.OrderQty.IsSet() {
		localVarPostBody["orderQty"] = localVarOptionals.OrderQty.Value()
	}

	if localVarOptionals != nil && localVarOptionals.Price.IsSet() {
		localVarPostBody["price"] = localVarOptionals.Price.Value()
	}

	if localVarOptionals != nil && localVarOptionals.DisplayQty.IsSet() {
		localVarPostBody["displayQty"] = localVarOptionals.DisplayQty.Value()
	}

	if localVarOptionals != nil && localVarOptionals.StopPx.IsSet() {
		localVarPostBody["stopPx"] = localVarOptionals.StopPx.Value()
	}

	if localVarOptionals != nil && localVarOptionals.ClOrdID.IsSet() {
		localVarPostBody["clOrdID"] = localVarOptionals.ClOrdID.Value()
	}

	if localVarOptionals != nil && localVarOptionals.ClOrdLinkID.IsSet() {
		localVarPostBody["clOrdLinkID"] = localVarOptionals.ClOrdLinkID.Value()
	}

	if localVarOptionals != nil && localVarOptionals.PegOffsetValue.IsSet() {
		localVarPostBody["pegOffsetValue"] = localVarOptionals.PegOffsetValue.Value()
	}

	if localVarOptionals != nil && localVarOptionals.PegPriceType.IsSet() {
		localVarPostBody["pegPriceType"] = localVarOptionals.PegPriceType.Value()
	}

	if localVarOptionals != nil && localVarOptionals.OrdType.IsSet() {
		localVarPostBody["ordType"] = localVarOptionals.OrdType.Value()
	}

	if localVarOptionals != nil && localVarOptionals.TimeInForce.IsSet() {
		localVarPostBody["timeInForce"] = localVarOptionals.TimeInForce.Value()
	}

	if localVarOptionals != nil && localVarOptionals.ExecInst.IsSet() {
		localVarPostBody["execInst"] = localVarOptionals.ExecInst.Value()
	}

	if localVarOptionals != nil && localVarOptionals.ContingencyType.IsSet() {
		localVarPostBody["contingencyType"] = localVarOptionals.ContingencyType.Value()
	}

	if localVarOptionals != nil && localVarOptionals.Text.IsSet() {
		localVarPostBody["text"] = localVarOptionals.Text.Value()
	}

	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
			var key string
			if auth.Prefix != "" {
				key = auth.Prefix + " " + auth.Key
			} else {
				key = auth.Key
			}
			localVarHeaderParams["api-key"] = key

		}
	}

	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))

		return localVarReturnValue, localVarHTTPResponse, err
	}

	newErr := GenericSwaggerError{
		body:  localVarBody,
		error: localVarHTTPResponse.Status,
	}

	if localVarHTTPResponse.StatusCode >= 400 && localVarHTTPResponse.StatusCode < 500 {
		var v ModelError
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
		} else {
			newErr.model = v
		}
	}

	return localVarReturnValue, localVarHTTPResponse, newErr
}

// OrderNewBulkOpts optinal args for create multiple new orders.
type OrderNewBulkOpts struct {
	Orders optional.String
}

// OrderNewBulk create multiple new orders for the same symbol.
func (a *OrderAPIService) OrderNewBulk(ctx context.Context, localVarOptionals *OrderNewBulkOpts) ([]Order, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("Post")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue []Order
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/order/bulk"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json", "application/x-www-form-urlencoded"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json", "application/xml", "text/xml", "application/javascript", "text/javascript"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.Orders.IsSet() {
		localVarFormParams.Add("orders", parameterToString(localVarOptionals.Orders.Value(), ""))
	}
	if ctx != nil {
		// API Key Authentication
		if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
			var key string
			if auth.Prefix != "" {
				key = auth.Prefix + " " + auth.Key
			} else {
				key = auth.Key
			}
			localVarHeaderParams["api-key"] = key

		}
	}
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-nonce"] = key

	// 	}
	// }
	// if ctx != nil {
	// 	// API Key Authentication
	// 	if auth, ok := ctx.Value(ContextAPIKey).(APIKey); ok {
	// 		var key string
	// 		if auth.Prefix != "" {
	// 			key = auth.Prefix + " " + auth.Key
	// 		} else {
	// 			key = auth.Key
	// 		}
	// 		localVarHeaderParams["api-signature"] = key

	// 	}
	// }
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHTTPResponse, err
		}
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}

		if localVarHTTPResponse.StatusCode == 200 {
			var v []Order
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 400 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 401 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		if localVarHTTPResponse.StatusCode == 404 {
			var v ModelError
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
