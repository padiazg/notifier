package utils

import (
	"crypto/rand"
	"encoding/hex"
)

type IdLenth uint

const (
	ID8  IdLenth = 8
	ID12 IdLenth = 12
	ID16 IdLenth = 16
)

var randRead = rand.Read

func RandomId(count IdLenth) string {
	var bytes = make([]byte, count)

	if _, err := randRead(bytes); err != nil {
		panic(err)
	}

	return hex.EncodeToString(bytes)
}
