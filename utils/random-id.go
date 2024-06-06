package utils

import (
	"crypto/rand"
	"encoding/hex"
)

var randRead = rand.Read

func RandomId8() string {
	var bytes = make([]byte, 8)

	if _, err := randRead(bytes); err != nil {
		panic(err)
	}

	return hex.EncodeToString(bytes)
}
