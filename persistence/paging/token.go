package paging

import (
	"encoding/base64"
	"errors"

	"github.com/caarlos0/env/v11"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/tesserical/geck/security/cryptox"
)

// TokenConfig represents the configuration for page tokens.
type TokenConfig struct {
	CipherKey string `env:"PAGE_TOKEN_CIPHER_KEY" envDefault:"HjwgM,F4?cA3t5Z6"`

	CipherKeyBytes []byte `env:"-"`
}

// NewTokenConfig creates a new token configuration.
func NewTokenConfig() (TokenConfig, error) {
	config, err := env.ParseAs[TokenConfig]()
	if err != nil {
		return TokenConfig{}, err
	}

	notValid := len(config.CipherKey) != 16 && len(config.CipherKey) != 24 && len(config.CipherKey) != 32
	if notValid {
		return TokenConfig{}, errors.New("invalid page token cipher key length, must be 16, 24 or 32 bytes")
	}
	config.CipherKeyBytes = []byte(config.CipherKey)
	return config, nil
}

// NewToken creates a new token from the given value.
// The value `v` represents the query parameters the caller is executing against its persistence storage.
//
// Moreover, tokens should replace any query parameters if present to avoid inconsistencies when querying
// pages using tokens.
//
// It is important to make sure that `v` is serializable (fields are exported and serializable as well).
//
// Use [ParseToken] to parse the token back into the value.
func NewToken(cipherKey []byte, v any) (string, error) {
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
func ParseToken(cipherKey []byte, encoded string, v any) error {
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
