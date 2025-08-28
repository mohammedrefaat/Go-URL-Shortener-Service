package utils

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func EncodeBase62(num uint64) string {
	if num == 0 {
		return "0"
	}

	var result []byte
	base := uint64(len(base62Chars))

	for num > 0 {
		result = append([]byte{base62Chars[num%base]}, result...)
		num = num / base
	}

	return string(result)
}

func DecodeBase62(str string) uint64 {
	var result uint64
	base := uint64(len(base62Chars))

	for _, char := range str {
		var value uint64
		switch {
		case char >= '0' && char <= '9':
			value = uint64(char - '0')
		case char >= 'A' && char <= 'Z':
			value = uint64(char - 'A' + 10)
		case char >= 'a' && char <= 'z':
			value = uint64(char - 'a' + 36)
		default:
			return 0 // Invalid character
		}
		result = result*base + value
	}

	return result
}
