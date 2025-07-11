package utilsCrypto

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/xid"
)

var (
	ErrInvalidTokenFormat   = errors.New("invalid token format")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrInvalidPublicKey     = errors.New("invalid public key")
	ErrTokenExpired         = errors.New("token has expired")
	ErrTokenNotValidYet     = errors.New("token not valid yet")
	ErrInvalidSignature     = errors.New("invalid token signature")
)

func BuildJwtSignerEd25519(ed25519PrivateKey, ed25519PublicKey string) (JwtSignerEd25519, error) {
	privateKeyInstance, err := decodePrivateKeyEd25519FromString(ed25519PrivateKey)
	if err != nil {
		return JwtSignerEd25519{}, err
	}
	publicKeyInstance, err := decodePublicKeyEd25519(ed25519PublicKey)
	if err != nil {
		return JwtSignerEd25519{}, err
	}
	return JwtSignerEd25519{privateKey: privateKeyInstance, publicKey: publicKeyInstance}, nil
}

type JwtSignerEd25519 struct {
	publicKey  *ed25519.PublicKey
	privateKey *ed25519.PrivateKey
}

func (j *JwtSignerEd25519) GenerateAndSignJwtTokenEd25519(data []AgKeyValue, expire time.Time) (string, error) {
	return generateAndSignEd25519JwtToken(data, expire, j.privateKey)
}

func (j *JwtSignerEd25519) VerifySignatureJwtTokenEd25519(token string) (*jwt.Token, error) {
	return verifySignatureJwtTokenEd25519(token, j.publicKey)
}

func generateAndSignEd25519JwtToken(data []AgKeyValue, expire time.Time, privateKey *ed25519.PrivateKey) (string, error) {
	claims := make(jwt.MapClaims)
	nounce := xid.New().String()

	claims["nounce"] = nounce
	claims["exp"] = expire.Unix()
	claims["stsession"] = time.Now().Unix()

	for _, itemData := range data {
		claims[itemData.Key] = itemData.Value
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(*privateKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func verifySignatureJwtTokenEd25519(token string, publicKey *ed25519.PublicKey) (*jwt.Token, error) {
	if token == "" {
		return nil, ErrInvalidTokenFormat
	}
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			//return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
			return nil, ErrInvalidSigningMethod
		}
		if publicKey == nil {
			return nil, ErrInvalidPublicKey
		}
		return *publicKey, nil
	})
	// if err != nil {
	// 	return nil, err
	// }
	if err != nil {
		// Handle specific validation errors
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				return nil, ErrTokenExpired
			case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
				return nil, ErrTokenNotValidYet
			case ve.Errors&jwt.ValidationErrorSignatureInvalid != 0:
				return nil, ErrInvalidSignature
			default:
				return nil, fmt.Errorf("token validation error: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	return parsedToken, nil
}

var (
	ErrBadPrivateKeySize         = fmt.Errorf("bad ed25519 private key size")
	ErrBadPublicKeySize          = fmt.Errorf("bad ed25519 public key size")
	ErrBadPrivateKeyBase64Format = fmt.Errorf("bad ed25519 private key base64 format")
)

func newEd25519PrivateKey(privateKeyBytes []byte) (*ed25519.PrivateKey, error) {
	if len(privateKeyBytes) != 32 && len(privateKeyBytes) != 64 {
		return &ed25519.PrivateKey{}, ErrBadPrivateKeySize
	}
	privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)
	return &privateKey, nil
}

func newEd25519PublicKey(publicKeyBytes []byte) (*ed25519.PublicKey, error) {
	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return &ed25519.PublicKey{}, ErrBadPrivateKeySize
	}
	publicKey := ed25519.PublicKey(publicKeyBytes)
	return &publicKey, nil
}

func decodePrivateKeyEd25519FromString(privateKeyString string) (*ed25519.PrivateKey, error) {
	decodedPrivateKey, err := hex.DecodeString(privateKeyString)
	if err != nil {
		return &ed25519.PrivateKey{}, err
	}
	key, err := newEd25519PrivateKey(decodedPrivateKey)
	if err != nil {
		return &ed25519.PrivateKey{}, err
	}
	return key, nil
}

func decodePublicKeyEd25519(publicKeyString string) (*ed25519.PublicKey, error) {
	decodedPublicKey, err := hex.DecodeString(publicKeyString)
	if err != nil {
		return &ed25519.PublicKey{}, err
	}
	publicKey, err := newEd25519PublicKey(decodedPublicKey)
	if err != nil {
		return &ed25519.PublicKey{}, err
	}
	return publicKey, nil
}
