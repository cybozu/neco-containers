package authenticator

import "testing"

func TestParseMode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Mode
	}{
		{
			name:  "x509 lowercase",
			input: "x509",
			want:  ModeX509,
		},
		{
			name:  "x509 uppercase",
			input: "X509",
			want:  ModeX509,
		},
		{
			name:  "jwt lowercase",
			input: "jwt",
			want:  ModeJWT,
		},
		{
			name:  "jwt uppercase",
			input: "JWT",
			want:  ModeJWT,
		},
		{
			name:  "jwt mixed case",
			input: "Jwt",
			want:  ModeJWT,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "invalid mode",
			input: "invalid",
			want:  "",
		},
		{
			name:  "similar but invalid",
			input: "x5099",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMode(tt.input)
			if got != tt.want {
				t.Errorf("ParseMode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
