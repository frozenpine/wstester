package pkcs8

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// PublicKey public key struct
type PublicKey struct {
	keyString string
	Key       *rsa.PublicKey
	block     *pem.Block
}

// Pem public key pem string
func (key *PublicKey) Pem() string {
	return key.keyString
}

// Type public key type
func (key *PublicKey) Type() string {
	return key.block.Type
}

// Headers public key headers
func (key *PublicKey) Headers() map[string]string {
	return key.block.Headers
}

// Bytes public key der bytes
func (key *PublicKey) Bytes() []byte {
	return key.block.Bytes
}

func (key *PublicKey) Encrypt(msg string) string {
	encrypted, _ := rsa.EncryptPKCS1v15(rand.Reader, key.Key, []byte(msg))
	return base64.StdEncoding.EncodeToString(encrypted)
}

// ParseFromPublicKeyString Parse public key string
func ParseFromPublicKeyString(keyString string, format PEMFormat) (PublicKey, error) {
	pub := PublicKey{}

	missingPadding := len(keyString) % 4
	if missingPadding > 0 {
		keyString = keyString + strings.Repeat("=", 4-missingPadding)
	}

	switch format {
	case PKCS1:
		keyString = fmt.Sprintf("-----BEGIN RSA PUBLIC KEY-----\n%s\n-----END RSA PUBLIC KEY-----", keyString)
	case PKCS8:
		keyString = fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", keyString)
	default:
		return pub, errors.New("invalid format")
	}

	block, rest := pem.Decode([]byte(keyString))

	if block == nil || len(rest) > 0 {
		return pub, errors.New("invalid pub key string")
	}

	// PKCS#8格式的公钥在解析时会报 "x509: trailing data after ASN.1 of public-key" 错误
	// 剪切多余字节, 不影响公钥的生成, 且已验证正确性
	var trailing int
	if format == PKCS8 {
		trailing = len(block.Bytes) - 76
	} else {
		trailing = len(block.Bytes)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes[:trailing])
	if err != nil {
		return pub, err
	}

	pub.keyString = keyString
	pub.block = block
	pub.Key = key.(*rsa.PublicKey)

	return pub, nil
}
