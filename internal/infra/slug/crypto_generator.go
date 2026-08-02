package slug

import (
	"crypto/rand"
	"encoding/binary"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const DefaultSlugLength = 7

type CryptoGenerator struct {
	length int
}

func NewCryptoGenerator(length int) *CryptoGenerator {
	if length == 0 {
		length = DefaultSlugLength
	}
	return &CryptoGenerator{length: length}
}

func (g *CryptoGenerator) Generate() string {
	var num uint64
	err := binary.Read(rand.Reader, binary.BigEndian, &num)
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}

	if num == 0 {
		return string(base62Chars[0])
	}
	buf := make([]byte, 0, 11)
	for num > 0 {
		buf = append(buf, base62Chars[num%62])
		num /= 62
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	encoded := string(buf)

	if len(encoded) >= g.length {
		return encoded[:g.length]
	}

	padded := make([]byte, g.length)
	for i := 0; i < g.length-len(encoded); i++ {
		padded[i] = '0'
	}
	copy(padded[g.length-len(encoded):], encoded)
	return string(padded)
}
