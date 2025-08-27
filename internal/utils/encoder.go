package utils

import (
	"crypto/rand"
	"math/big"
)

const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func GenerateShortCode(length int) string {
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		result[i] = base62Chars[num.Int64()]
	}
	return string(result)
}

func EncodeBase62(num uint64) string {
	if num == 0 {
		return "0"
	}

	encoded := make([]byte, 0)
	base := uint64(len(base62Chars))

	for num > 0 {
		encoded = append([]byte{base62Chars[num%base]}, encoded...)
		num /= base
	}

	return string(encoded)
}

func DecodeBase62(encoded string) uint64 {
	decoded := uint64(0)
	base := uint64(len(base62Chars))

	for _, char := range encoded {
		for i, c := range base62Chars {
			if char == c {
				decoded = decoded*base + uint64(i)
				break
			}
		}
	}

	return decoded
}
