package output

import "testing"

func TestParseHyprlandTarget(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid address",
			input: `{"address":"0x1a2B"}`,
			want:  "0x1a2B",
		},
		{
			name:    "empty address",
			input:   `{"address":""}`,
			wantErr: true,
		},
		{
			name:    "invalid address",
			input:   `{"address":"window"}`,
			wantErr: true,
		},
		{
			name:    "invalid json",
			input:   `{`,
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseHyprlandTarget([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseHyprlandTarget() error = %v, want error: %t", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("parseHyprlandTarget() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseHyprlandBool(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{
			name:  "false",
			input: `{"int":0}`,
		},
		{
			name:  "true",
			input: `{"int":1}`,
			want:  true,
		},
		{
			name:    "invalid value",
			input:   `{"int":2}`,
			wantErr: true,
		},
		{
			name:    "invalid json",
			input:   `{`,
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseHyprlandBool([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseHyprlandBool() error = %v, want error: %t", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("parseHyprlandBool() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestIsX11Window(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name: "empty",
		},
		{
			name:  "numeric",
			input: "123",
			want:  true,
		},
		{
			name:  "alphanumeric",
			input: "12x",
		},
		{
			name:  "negative",
			input: "-123",
		},
		{
			name:  "whitespace",
			input: " 123 ",
		},
	}
	for _, tc := range cases {
		if got := isX11Window(tc.input); got != tc.want {
			t.Errorf("isX11Window(%q) = %t, want %t", tc.input, got, tc.want)
		}
	}
}

func TestPasteShortcut(t *testing.T) {
	t.Parallel()
	const target = "0x1a2b"
	const want = "CTRL,code:55,address:" + target
	if got := pasteShortcut(target); got != want {
		t.Errorf("pasteShortcut() = %q, want %q", got, want)
	}
}
