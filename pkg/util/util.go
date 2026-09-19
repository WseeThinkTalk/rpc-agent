package util

import (
	"crypto/rand"
	"math/big"
	"time"
)

func GenerateCode(size int) string {
	min := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(size-1)), nil)
	max := new(big.Int).Mul(big.NewInt(9), min)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return min.String()
	}
	code := new(big.Int).Add(min, n)
	return code.String()
}

func EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, t.Location())
}
