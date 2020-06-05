package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
)

func Authenticate(seed []byte, name, key string) []byte {
	if len(seed) != 16 {
		seed = make([]byte, 16)
		rand.Read(seed)
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(seed)
	h.Write([]byte(name))
	seed = append(seed, h.Sum(nil)...)
	return seed[:32]
}
