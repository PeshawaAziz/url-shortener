package slug

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/PeshawaAziz/url-shortener/pkg/encoding"
)

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
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}

	encoded := encoding.EncodeBase62(num)

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
