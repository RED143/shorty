package hash

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
)

func GenerateShortURL(originalURL string, baseURL string) (string, error) {
	hash := md5.Sum([]byte(originalURL))
	hashString := hex.EncodeToString(hash[:])

	shortURL, err := url.JoinPath(baseURL, hashString[:7])
	if err != nil {
		return "", fmt.Errorf("failed to generate shortURL: %w", err)
	}

	return shortURL, nil
}
