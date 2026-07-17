package encoding

import (
	"math"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func EncodeBase62(num uint64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	length := int(math.Log(float64(num))/math.Log(62)) + 1
	buf := make([]byte, length)

	for i := length - 1; i >= 0; i-- {
		buf[i] = base62Chars[num%62]
		num /= 62
	}

	return string(buf)
}

func DecodeBase62(s string) uint64 {
	var num uint64
	for _, c := range s {
		var val int
		switch {
		case '0' <= c && c <= '9':
			val = int(c - '0')
		case 'a' <= c && c <= 'z':
			val = int(c-'a') + 10
		case 'A' <= c && c <= 'Z':
			val = int(c-'A') + 36
		default:
			return 0
		}
		num = num*62 + uint64(val)
	}
	return num
}
