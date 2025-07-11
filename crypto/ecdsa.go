package utilsCrypto

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1"
)

func HexSignatureToDER(hexSignature string) ([]byte, error) {
	// Convert the hex string to bytes.
	signatureBytes, err := hex.DecodeString(hexSignature)
	if err != nil {
		return nil, err
	}

	// Ensure that the signature has the expected length (usually 64 bytes).
	if len(signatureBytes) != 64 {
		return nil, fmt.Errorf("InvalidSignatureLength")
	}

	// Split the signature into 'r' and 's' parts.
	rawR := signatureBytes[:32]
	rawS := signatureBytes[32:]

	// Parse the 'r' and 's' values as big integers.
	r := new(big.Int).SetBytes(rawR)
	s := new(big.Int).SetBytes(rawS)

	// Build the DER-encoded signature.
	derSignature, err := asn1.Marshal(struct {
		R, S *big.Int
	}{r, s})
	if err != nil {
		return nil, err
	}

	return derSignature, nil
}

func EcdsaSignatureVerify(hexPublicKey string, data []byte, hexSignature string) (bool, error) {
	publicKeyBytes, err := hex.DecodeString(hexPublicKey)
	if err != nil {
		return false, err
	}

	signatureBytes, err := HexSignatureToDER(hexSignature)
	if err != nil {
		return false, err
	}

	// Generate the SHA-256 hash of the text string
	hash := sha256.Sum256(data)

	// Parse the public key
	publicKey, err := secp256k1.ParsePubKey(publicKeyBytes)
	if err != nil {
		return false, err
	}

	signature, err := secp256k1.ParseSignature(signatureBytes)
	if err != nil {
		return false, err
	}

	// Verify the signature
	verified := signature.Verify(hash[:], publicKey)

	return verified, nil
}
