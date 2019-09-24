package ngerest

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/frozenpine/pkcs8"
)

var (
	_ context.Context
)

// KeyExchangeService public key exchange service for NGE
type KeyExchangeService service

// GetPublicKey get public key from nge
func (a *KeyExchangeService) GetPublicKey(ctx context.Context) (pkcs8.PublicKey, *http.Response, error) {
	var (
		localVarHTTPMethod  = strings.ToUpper("POST")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue pkcs8.PublicKey
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/user/getPublicKey"

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

	newErr := GenericSwaggerError{
		body:  localVarBody,
		error: localVarHTTPResponse.Status,
	}

	if localVarHTTPResponse.StatusCode < 300 {
		data := make(map[string]interface{})
		err := json.Unmarshal(localVarBody, &data)

		if err != nil {
			newErr.error = err.Error()
			newErr.body = localVarBody
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		pubKeyString, ok := data["result"].(string)
		if !ok {
			newErr.error = "invalid result type in response"
			newErr.body = localVarBody
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		localVarReturnValue, err := pkcs8.ParseFromPublicKeyString(pubKeyString, pkcs8.PKCS8)
		if err != nil {
			newErr.error = err.Error()
			newErr.body = localVarBody
			return localVarReturnValue, localVarHTTPResponse, newErr
		}

		return localVarReturnValue, localVarHTTPResponse, err
	}

	newErr.error = "request failed"
	newErr.body = localVarBody
	return localVarReturnValue, localVarHTTPResponse, newErr
}
