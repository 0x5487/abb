package app

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func sha256Hash(text string, encodingType string) string {
	rawBytes := []byte(text)
	h := sha256.Sum256(rawBytes)
	result := ""
	if encodingType == "base64" {
		result = base64.StdEncoding.EncodeToString(h[:])
	} else {
		result = hex.EncodeToString(h[:])
	}

	return result
}

func SHA256EncodeToBase64(text string) string {
	return sha256Hash(text, "base64")
}
