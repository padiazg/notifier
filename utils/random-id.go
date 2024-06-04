package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func RamdomId8() string {
	var bytes = make([]byte, 8)

	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}

	return hex.EncodeToString(bytes)
}
