package security

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2Hasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{time: 1, memory: 64 * 1024, threads: 2, keyLen: 32}
}

func (h *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, 16)
	hash := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)
	b64Hash := base64.StdEncoding.EncodeToString(hash)
	b64Salt := base64.StdEncoding.EncodeToString(salt)
	return "argon2id$v=19$m=65536,t=1,p=2$" + b64Salt + "$" + b64Hash, nil
}

func (h *Argon2Hasher) Compare(encodedHash, password string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 {
		return false
	}
	salt, _ := base64.StdEncoding.DecodeString(parts[2])
	hash, _ := base64.StdEncoding.DecodeString(parts[3])

	compareHash := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)
	return subtle.ConstantTimeCompare(hash, compareHash) == 1
}
