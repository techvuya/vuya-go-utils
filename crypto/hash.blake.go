package utilsCrypto

import (
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

func getBlake256Hash(data []byte) string {
	hash := blake2b.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func verifyDataBlake256Hash(data []byte, hashHexString string) bool {
	hash := getBlake256Hash(data)
	return hash == hashHexString
}
