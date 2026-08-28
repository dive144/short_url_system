package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateRandomShortCode(
	length int,
) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf(
			"短码长度必须大于0",
		)
	}

	const characters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	result := make([]byte, length)

	max := big.NewInt(
		int64(len(characters)),
	)

	for i := range result {
		randomIndex, err := rand.Int(
			rand.Reader,
			max,
		)
		if err != nil {
			return "", fmt.Errorf(
				"生成随机短码失败: %w",
				err,
			)
		}

		index := randomIndex.Int64()
		result[i] = characters[index]
	}

	return string(result), nil
}
