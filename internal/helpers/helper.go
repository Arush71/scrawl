package helpers

import "math/rand"

const (
	charset       string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength int    = len(charset)
)

func RandomStr() string {
	bd := make([]byte, 6)
	for i := range 6 {
		bd[i] = charset[rand.Intn(charsetLength)]
	}
	return string(bd)
}
