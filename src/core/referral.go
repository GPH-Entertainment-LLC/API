package core

import (
	"crypto/rand"
	"math/big"
	"strings"
)

func GenerateReferralCode() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 20
	var result strings.Builder

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result.WriteByte(charset[randomIndex.Int64()])
	}

	return result.String(), nil
}
