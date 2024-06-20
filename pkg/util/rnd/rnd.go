// Package rnd has utility functions for randomization
package rnd

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
)

// RandomID generates a random id based on prefix and specified length
//
// output will always be "<prefix>-<chars_of_length>"
func RandomID(prefix string, length uint) string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("%s-", prefix))

	for i := uint(0); i < length; i++ {
		nbig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(RandomIDPool))))

		val := nbig.Int64()
		byteR := RandomIDPool[val]
		buffer.WriteByte(byteR)
	}

	return buffer.String()
}

// RandomString generates a random string based on length
//
// output will always be "<chars_of_length>"
func RandomString(length uint) string {
	var buffer bytes.Buffer

	for i := uint(0); i < length; i++ {
		nbig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(RandomIDPool))))

		val := nbig.Int64()
		byteR := RandomIDPool[val]
		buffer.WriteByte(byteR)
	}

	return buffer.String()
}
