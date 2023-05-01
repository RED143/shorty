package hash

import (
	"crypto/md5"
	"encoding/hex"
)

func Generate(data []byte) string {
	hash := md5.Sum(data)
	hashString := hex.EncodeToString(hash[:])

	return hashString[:7]
}
