package validation

import (
	"encoding"
	"errors"
	"fmt"
)

// - Codec -

var (
	ErrInvalidCodecDriver = errors.New("validation: codec driver is invalid")
)

// CodecDriver is the codec algorithm to use for struct tag field scanning.
//
// As validators scan struct fields, some of them might define tags for specific codecs (e.g. json).
// This causes a mismatch between the base and the encoded names and lead to confusion for routine callers.
//
// A CodecDriver defines which tags of a codec mechanism to use for validator operations
// like error generation.
type CodecDriver uint8

const (
	// JSONDriver is the JSON struct field driver.
	//
	// Uses `json` field tag.
	JSONDriver = CodecDriver(iota) + 1
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
	// EnvironmentDriver is the environment struct field driver.
	//
	// Uses `env` field tag.
	EnvironmentDriver
)

var (
	// compile-time assertion
	_ fmt.Stringer             = CodecDriver(0)
	_ encoding.TextUnmarshaler = (*CodecDriver)(nil)
	_ encoding.TextMarshaler   = (*CodecDriver)(nil)

	_structFieldFromStringMap = map[string]CodecDriver{
		"json": JSONDriver,
		"yaml": YAMLDriver,
		"xml":  XMLDriver,
		"toml": TOMLDriver,
		"env":  EnvironmentDriver,
	}
	_structFieldToStringMap = map[CodecDriver]string{
		JSONDriver:        "json",
		YAMLDriver:        "yaml",
		XMLDriver:         "xml",
		TOMLDriver:        "toml",
		EnvironmentDriver: "env",
	}
)

// ParseCodecDriver allocates a new [CodecDriver] instance based on its string value.
func ParseCodecDriver(v string) (CodecDriver, error) {
	if d, ok := _structFieldFromStringMap[v]; ok {
		return d, nil
	}
	return 0, ErrInvalidCodecDriver
}

func (d CodecDriver) String() string {
	return _structFieldToStringMap[d]
}

func (d CodecDriver) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

func (d *CodecDriver) UnmarshalText(text []byte) error {
	v, err := ParseCodecDriver(string(text))
	if err != nil {
		return err
	}
	*d = v
	return nil
}
