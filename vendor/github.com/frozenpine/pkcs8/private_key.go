package pkcs8

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
)

type PrivateKey struct {
	*rsa.PrivateKey
}

func (key *PrivateKey) GetPublicKey() string {
	der, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)

	return base64.StdEncoding.EncodeToString(der)
}

func (key *PrivateKey) Decrypt(msg string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return "", err
	}

	decrypt, err := rsa.DecryptPKCS1v15(rand.Reader, key.PrivateKey, cipherBytes)
	return string(decrypt), err
}

func GeneratePriveKey(bit int) PrivateKey {
	pri := PrivateKey{}

	key, _ := rsa.GenerateKey(rand.Reader, bit)

	pri.PrivateKey = key

	return pri
}
