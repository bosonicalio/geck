package paging

import (
	"encoding/base64"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/tesserical/geck/security/cryptox"
)

// - Cipher -

// TokenCipherKey is a type alias for a byte slice representing the cipher key used for encrypting and decrypting tokens.
// It is used to ensure that the cipher key is of the correct type when passed to functions that handle token
// encryption and decryption.
//
// The cipher key must be 16, 24, or 32 bytes long, depending on the encryption algorithm used.
//
// A separate type is defined so developers can easily inject the cipher key into the application
// configuration, ensuring that it is used consistently across the application.
type TokenCipherKey []byte

// - Factory/Parser -

// NewToken creates a new token from the given value.
// The value `v` represents the query parameters the caller is executing against its persistence storage.
//
// Moreover, tokens should replace any query parameters if present to avoid inconsistencies when querying
// pages using tokens.
//
// It is important to make sure that `v` is serializable (fields are exported and serializable as well).
//
// Use [ParseToken] to parse the token back into the value.
func NewToken(cipherKey TokenCipherKey, v any) (string, error) {
	serialized, err := msgpack.Marshal(v)
	if err != nil {
		return "", err
	}

	encrypted, err := cryptox.Encrypt(serialized, cipherKey)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(encrypted), nil
}

// ParseToken parses the given token into the given value.
//
// Use [NewToken] to create a token from a value.
func ParseToken(cipherKey TokenCipherKey, encoded string, v any) error {
	encrypted, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}

	serialized, err := cryptox.Decrypt(encrypted, cipherKey)
	if err != nil {
		return err
	}

	return msgpack.Unmarshal(serialized, v)
}
