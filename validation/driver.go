package validation

import (
	"encoding"
	"errors"
	"fmt"
)

var (
	ErrInvalidDriver = errors.New("validation: driver is invalid")
)

// Driver refers to the underlying engine (or third-party package) used for validation.
type Driver uint8

const (
	// GoPlaygroundDriver is the [github.com/go-playground/validator/v10] driver.
	GoPlaygroundDriver = Driver(iota) + 1
)

var (
	// compile-time assertion
	_ fmt.Stringer             = Driver(0)
	_ encoding.TextUnmarshaler = (*Driver)(nil)
	_ encoding.TextMarshaler   = (*Driver)(nil)

	_driverFromStringMap = map[string]Driver{
		"go-playground": GoPlaygroundDriver,
	}
	_driverToStringMap = map[Driver]string{
		GoPlaygroundDriver: "go-playground",
	}
)

// ParseDriver allocates a new [Driver] instance based on its string value.
func ParseDriver(v string) (Driver, error) {
	if d, ok := _driverFromStringMap[v]; ok {
		return d, nil
	}
	return 0, ErrInvalidDriver
}

// String returns the string representation of the driver.
func (d Driver) String() string {
	return _driverToStringMap[d]
}

// MarshalText implements the [encoding.TextMarshaler] interface for Driver.
func (d Driver) MarshalText() (text []byte, err error) {
	if str, ok := _driverToStringMap[d]; ok {
		return []byte(str), nil
	}
	return nil, ErrInvalidDriver
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface for Driver.
func (d *Driver) UnmarshalText(text []byte) error {
	v, err := ParseDriver(string(text))
	if err != nil {
		return err
	}
	*d = v
	return nil
}

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
