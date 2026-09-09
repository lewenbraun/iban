package app

import (
	"testing"

	"github.com/lewenbraun/iban/iban/internal/state"
)

func TestOptionsMode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want state.Mode
	}{
		{
			name: "copy",
			want: state.ModeCopy,
		},
		{
			name: "paste",
			args: []string{"-paste"},
			want: state.ModePaste,
		},
		{
			name: "paste and enter",
			args: []string{"-enter"},
			want: state.ModePasteEnter,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts, err := parseOptions(tc.args)
			if err != nil {
				t.Fatalf("parseOptions() error = %v", err)
			}
			if got := opts.mode(); got != tc.want {
				t.Errorf("mode() = %q, want %q", got, tc.want)
			}
		})
	}
}
