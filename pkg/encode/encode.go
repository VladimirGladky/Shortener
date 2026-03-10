package encode

const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	idOffset    = 100000
)

func EncodeBase62(num int) string {
	num += idOffset

	if num == 0 {
		return string(base62Chars[0])
	}

	result := ""
	base := len(base62Chars)

	for num > 0 {
		remainder := num % base
		result = string(base62Chars[remainder]) + result
		num = num / base
	}

	return result
}
