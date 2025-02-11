package pagetoken

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"
	"github.com/vmihailenco/msgpack/v5"
)

// Token is an informational structure used as mechanism to fetch specific segments of datasets.
//
// For a page [Token] to work, dataset MUST be sorted by a `cursor` (order does matter) so it can
// indicate where to start fetching items for the new page. Depending on the [Direction], [Token] processors
// will (and must) use the respective cursor value (StartCursor, EndCursor).
//
// It is recommended [Token] instances are processed by low-level components (i.e. persistence storage) as
// every dataset source is different and how a [Token] is handled can change.
type Token struct {
	// Name of the cursor used to paginate the dataset.
	CursorName string
	// Value of a cursor. Must be populated if Direction is [PreviousDirection], nil otherwise.
	StartCursor any
	// Value of a cursor. Must be populated if Direction is [NextDirection], nil otherwise.
	EndCursor any
	// Sorting specification.
	Sort Sort
	// Indicates what page to fetch (previous or next).
	Direction Direction
}

// Sort is an informational structure to specify the sorting mechanism of a [Token].
type Sort struct {
	Field    string
	Operator string
}

// Direction is an informational type to specify the direction of a [Token] ([NextDirection], [PreviousDirection]).
type Direction string

const (
	// NextDirection specifies to fetch the next page of a dataset.
	NextDirection Direction = "next"
	// PreviousDirection specifies to fetch the previous page of a dataset.
	PreviousDirection Direction = "previous"
)

var (
	// compile-time assertions
	_ fmt.Stringer   = (*Token)(nil)
	_ json.Marshaler = (*Token)(nil)
)

// MarshalJSON encodes and encrypts a [Token], to later finally encode it as a JSON string.
func (t *Token) MarshalJSON() ([]byte, error) {
	encodedVal, err := Marshal(t)
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(encodedVal))
}

func (t *Token) String() string {
	val, err := Marshal(t)
	if err != nil {
		return ""
	}
	return string(val)
}

// Marshal encrypts a `v` ([Token]) to later encode it as base64 (url-safe).
func Marshal(v *Token, opts ...TokenOption) ([]byte, error) {
	tokenMsgpack, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}

	options := tokenOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	encryptionKey := lo.CoalesceOrEmpty(options.encryptionKey, defaultEncryptionKey)
	encryptedToken, err := encrypt(tokenMsgpack, []byte(encryptionKey))
	if err != nil {
		return nil, err
	}
	encodedToken := make([]byte, base64.URLEncoding.EncodedLen(len(encryptedToken)))
	base64.URLEncoding.Encode(encodedToken, encryptedToken)
	return encodedToken, nil
}

// UnmarshalEmptyable unmarshals `encodedToken` if given value is not empty.
func UnmarshalEmptyable(encodedToken string, opts ...TokenOption) (*Token, error) {
	if len(encodedToken) == 0 {
		return nil, nil
	}
	return Unmarshal(encodedToken, opts...)
}

// Unmarshal decodes `encodedToken` (encoded [Token]) (from base64, url-safe) to later decrypt it and
// parse it into [Token].
func Unmarshal(encodedToken string, opts ...TokenOption) (*Token, error) {
	encryptedToken, err := base64.URLEncoding.DecodeString(encodedToken)
	if err != nil {
		return nil, err
	}

	options := tokenOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	encryptionKey := lo.CoalesceOrEmpty(options.encryptionKey, defaultEncryptionKey)
	tokenMsgpack, err := decrypt(encryptedToken, []byte(encryptionKey))
	if err != nil {
		return nil, err
	}
	var token Token
	err = msgpack.Unmarshal(tokenMsgpack, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// -- OPTIONS --

type tokenOptions struct {
	encryptionKey string
}

type TokenOption func(*tokenOptions)

func WithEncryptionKey(k string) TokenOption {
	return func(o *tokenOptions) {
		o.encryptionKey = k
	}
}
