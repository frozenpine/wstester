package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
)

// GenerateSignature generate signature for uri
func GenerateSignature(
	secret, method string, url *url.URL, expires int, body *bytes.Buffer) string {
	h := hmac.New(sha256.New, []byte(secret))

	path := url.Path
	if url.RawQuery != "" {
		path = path + "?" + url.RawQuery
	}

	var bodyString string
	if body != nil {
		bodyString = strings.TrimRight(body.String(), "\r\n")
	}

	message := strings.ToUpper(method) + path + strconv.Itoa(expires) + bodyString

	h.Write([]byte(message))

	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}
