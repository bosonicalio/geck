package version

import (
	"encoding"
	"errors"
	"fmt"

	"golang.org/x/mod/semver"
)

// Version represents a semantic version following the MAJOR.MINOR.PATCH convention.
// It is used to convey meaning about the changes introduced in a software release.
//
// Fields:
//
// - Major: Incremented when making incompatible API changes (e.g., breaking changes).
//
// - Minor: Incremented when adding functionality in a backward-compatible manner.
//
// - Patch: Incremented for backward-compatible bug fixes.
//
// Example:
//
// For a version "2.3.1":
//
// - Major = 2 (breaking changes introduced in version 2.0.0).
//
// - Minor = 3 (new features added since version 2.0.0).
//
// - Patch = 1 (bug fixes applied since version 2.3.0).
//
// This structure implements [encoding.TextMarshaler], [encoding.TextUnmarshaler] and
// [fmt.Stringer] for easier integration with external components.
type Version struct {
	raw        string
	major      string
	majorMinor string
	prerelease string
	build      string
	canonical  string
}

var (
	// compile-time assertions
	_ encoding.TextMarshaler   = (*Version)(nil)
	_ encoding.TextUnmarshaler = (*Version)(nil)
	_ fmt.Stringer             = (*Version)(nil)

	// ErrInvalidSemver the given string was not using the semantic version format.
	ErrInvalidSemver = errors.New("invalid semantic version")
)

// Parse allocates a new [Version] instance.
//
// Returns [ErrInvalidSemver] if `v` is an invalid semantic version.
func Parse(value string) (Version, error) {
	if ok := semver.IsValid(value); !ok {
		return Version{}, ErrInvalidSemver
	}
	return Version{
		raw:        value,
		major:      semver.Major(value),
		majorMinor: semver.MajorMinor(value),
		prerelease: semver.Prerelease(value),
		build:      semver.Build(value),
		canonical:  semver.Canonical(value),
	}, nil
}

// MustParse allocates a new [Version] instance and panics if the given string is not a valid semantic version.
func MustParse(value string) Version {
	ver, err := Parse(value)
	if err != nil {
		panic(fmt.Sprintf("version: %s is not a valid semantic version: %v", value, err))
	}
	return ver
}

func (v Version) MarshalText() (text []byte, err error) {
	return []byte(v.raw), nil
}

func (v *Version) UnmarshalText(text []byte) error {
	ver, err := Parse(string(text))
	if err != nil {
		return err
	}
	*v = ver
	return nil
}

func (v Version) String() string {
	return v.raw
}

// Major returns the major version prefix of the semantic version v.
// For example, Major("v2.1.0") == "v2".
// If v is an invalid semantic version string, Major returns the empty string.
func (v Version) Major() string {
	return v.major
}

// MajorMinor returns the major.minor version prefix of the semantic version v.
// For example, MajorMinor("v2.1.0") == "v2.1".
// If v is an invalid semantic version string, MajorMinor returns the empty string.
func (v Version) MajorMinor() string {
	return v.majorMinor
}

// Prerelease returns the prerelease suffix of the semantic version v.
// For example, Prerelease("v2.1.0-pre+meta") == "-pre".
// If v is an invalid semantic version string, Prerelease returns the empty string.
func (v Version) Prerelease() string {
	return v.prerelease
}

// Build returns the build suffix of the semantic version v.
// For example, Build("v2.1.0+meta") == "+meta".
// If v is an invalid semantic version string, Build returns the empty string.
func (v Version) Build() string {
	return v.build
}

// Canonical returns the canonical formatting of the semantic version v.
// It fills in any missing .MINOR or .PATCH and discards build metadata.
// Two semantic versions compare equal only if their canonical formattings
// are identical strings.
// The canonical invalid semantic version is the empty string.
func (v Version) Canonical() string {
	return v.canonical
}
