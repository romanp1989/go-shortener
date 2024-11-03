package utils

import (
	"crypto/rand"
	"encoding/binary"
)

func GenerateUniqueID(size int) (uint64, error) {
	randID := make([]byte, size)
	_, err := rand.Read(randID)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(randID), nil
}
