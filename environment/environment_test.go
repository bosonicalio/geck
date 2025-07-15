package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Environment
		wantErr bool
	}{
		{
			name:    "Should return Production for 'production'",
			in:      "production",
			want:    Production,
			wantErr: false,
		},
		{
			name:    "Should return Production for 'prod'",
			in:      "prod",
			want:    Production,
			wantErr: false,
		},
		{
			name:    "Should return Staging for 'staging'",
			in:      "staging",
			want:    Staging,
			wantErr: false,
		},
		{
			name:    "Should return Staging for 'stage'",
			in:      "stage",
			want:    Staging,
			wantErr: false,
		},
		{
			name:    "Should return Staging for 'stg'",
			in:      "stg",
			want:    Staging,
			wantErr: false,
		},
		{
			name:    "Should return Staging for 'sandbox'",
			in:      "sandbox",
			want:    Staging,
			wantErr: false,
		},
		{
			name:    "Should return Staging for 'snx'",
			in:      "snx",
			want:    Staging,
			wantErr: false,
		},
		{
			name:    "Should return Staging for 'pilot'",
			in:      "pilot",
			want:    Staging,
			wantErr: false,
		},
		{
			name:    "Should return Development for 'development'",
			in:      "development",
			want:    Development,
			wantErr: false,
		},
		{
			name:    "Should return Development for 'dev'",
			in:      "dev",
			want:    Development,
			wantErr: false,
		},
		{
			name:    "Should return Local for 'local'",
			in:      "local",
			want:    Local,
			wantErr: false,
		},
		{
			name:    "Should return error for unknown environment",
			in:      "unknown",
			want:    Unknown,
			wantErr: true,
		},
		{
			name:    "Should return error for empty environment",
			in:      "",
			want:    Unknown,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.in)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
