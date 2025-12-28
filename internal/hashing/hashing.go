package hashing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func Hash(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil)), nil
}
