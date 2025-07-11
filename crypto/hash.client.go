package utilsCrypto

import "github.com/rs/xid"

// HashService is a service that provides various cryptographic utilities,
// including password hashing using the Argon2id algorithm, data hashing (Blake2b-256),
// and token generation. The Argon2id parameters can be configured through
// the ArgonParams field of the HashService struct.
type HashService struct {
	argonParams *ArgonParams
}

// NewHashService creates a new instance of HashService with default Argon2id
// parameters for password hashing.
//
// Returns:
//
//	*HashService - A new instance of HashService to access cryptographic utilities with default Argon2id parameters.
func NewHashService() *HashService {
	return &HashService{argonParams: defaultPasswordHashParams}
}

// CreatePasswordHash generates a secure hash of the given password using the Argon2id algorithm.
// The hashing behavior is controlled by the Argon2id parameters configured in the HashService.
//
// Parameters:
//
//	password string - The password to be hashed.
//
// Returns:
//
//	string - The resulting hash of the password.
//	error - Returns an error if the hashing process fails.
func (c *HashService) CreatePasswordHash(password string) (string, error) {
	return createPasswordHashArgon2id(c.argonParams, password)
}

// ComparePasswordAndHash compares a given password with its Argon2id hash to verify if they match.
//
// Parameters:
//
//	password string - The plain-text password to compare.
//	hash string - The hashed password to compare against.
//
// Returns:
//
//	error - Returns nil if the password matches the hash. Otherwise, returns an error if they do not match.
func (c *HashService) ComparePasswordAndHash(password, hash string) error {
	return comparePasswordAndHashArgon2id(password, hash)
}

// CreateHashBlake256 creates a Blake2b-256 hash of the provided data.
//
// Parameters:
//
//	data []byte - The data to be hashed.
//
// Returns:
//
//	string - The hexadecimal string representation of the Blake2b-256 hash
func (c *HashService) CreateHashBlake256(data []byte) string {
	return getBlake256Hash(data)
}

// VerifyHashBlake256 verifies that the provided data matches the given Blake2b-256 hash.
//
// Parameters:
//
//	data []byte - The original data to be verified.
//	hashHexString string - The expected hexadecimal hash string.
//
// Returns:
//
//	bool - Returns true if the data matches the hash, otherwise returns false.
func (c *HashService) VerifyHashBlake256(data []byte, hashHexString string) bool {
	return verifyDataBlake256Hash(data, hashHexString)
}

// GenerateToken generates a unique token using the xid library, which generates
// globally unique IDs based on the current timestamp and machine-specific data.
//
// Returns:
//
//	string - A unique token string.
func (c *HashService) GenerateToken() string {
	return xid.New().String()
}
