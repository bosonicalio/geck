package version_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tesserical/geck/version"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
		wantErr bool
	}{
		{"Should be a valid version", "v1.0.0", "v1.0.0", false},
		{"Should be a valid version with prefix", "v2.3.4-alpha", "v2.3.4-alpha", false},
		{"Should return error with invalid version format", "1.0", "", true},
		{"Should return error with empty version", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := version.Parse(tt.version)
			assert.Equal(t, tt.wantErr, err != nil, "unexpected error status")
			assert.Equal(t, tt.want, got.String(), "unexpected version")
		})
	}
}
