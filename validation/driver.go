package validation

import (
	"encoding"
	"fmt"
)

// StructFieldDriver is the codec algorithm to use for struct tag field scanning.
//
// As validators scan struct fields, some of them might define tags for specific codecs (e.g. json).
// This causes a mismatch between the base and the encoded names and lead to confusion for routine callers.
//
// A StructFieldDriver defines which tags of a codec mechanism to use for validator operations
// like error generation.
type StructFieldDriver uint8

const (
	// JSONDriver is the JSON struct field driver.
	//
	// Uses `json` field tag.
	JSONDriver = StructFieldDriver(iota)
	// YAMLDriver is the YAML struct field driver.
	//
	// Uses `yaml` field tag.
	YAMLDriver
	// XMLDriver is the XML struct field driver.
	//
	// Uses `xml` field tag.
	XMLDriver
	// TOMLDriver is the TOML struct field driver.
	//
	// Uses `toml` field tag.
	TOMLDriver
)

var (
	// compile-time assertion
	_ fmt.Stringer             = StructFieldDriver(0)
	_ encoding.TextUnmarshaler = (*StructFieldDriver)(nil)
	_ encoding.TextMarshaler   = (*StructFieldDriver)(nil)

	_structFieldInternalStringMap = map[string]StructFieldDriver{
		"json": JSONDriver,
		"yaml": YAMLDriver,
		"xml":  XMLDriver,
		"toml": TOMLDriver,
	}
	_structFieldDriverStringMap = map[StructFieldDriver]string{
		JSONDriver: "json",
		YAMLDriver: "yaml",
		XMLDriver:  "xml",
		TOMLDriver: "toml",
	}
)

// NewStructFieldDriver allocates a new [StructFieldDriver] instance based on its string value.
//
// Default value is [JSONDriver].
func NewStructFieldDriver(v string) StructFieldDriver {
	return _structFieldInternalStringMap[v]
}

func (d StructFieldDriver) String() string {
	return _structFieldDriverStringMap[d]
}

func (d StructFieldDriver) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

func (d *StructFieldDriver) UnmarshalText(text []byte) error {
	*d = NewStructFieldDriver(string(text))
	return nil
}
