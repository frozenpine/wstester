package pkcs8

type PEMFormat string

const (
	// PKCS1 pkcs#1
	PKCS1 PEMFormat = "pkcs#1"
	// PKCS8 pkcs#8
	PKCS8 PEMFormat = "pkcs#8"
)
