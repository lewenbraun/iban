package dictation

import (
	"testing"
	"time"
)

func TestSplitPasteCommand(t *testing.T) {
	t.Parallel()
	const helloWorld = "hello world"
	cases := []struct {
		name  string
		input string
		want  string
		paste bool
	}{
		{"plain text", helloWorld, helloWorld, false},
		{"english paste", helloWorld + " paste", helloWorld, true},
		{"english with bang", "do it, paste!", "do it", true},
		{"russian suffix", "привет коля вставь", "привет коля", true},
		{"russian imperative", "отправь письмо вставляй", "отправь письмо", true},
		{"russian polite", "запиши вставьте", "запиши", true},
		{"only command", "вставь", "", true},
		{"embedded not suffix", "pasted text", "pasted text", false},
		{"russian word contains", "вставка", "вставка", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, paste := SplitPasteCommand(tc.input)
			if got != tc.want || paste != tc.paste {
				t.Errorf("SplitPasteCommand(%q) = (%q, %v), want (%q, %v)",
					tc.input, got, paste, tc.want, tc.paste)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("truncate short = %q", got)
	}
	if got := truncate("привет мир как дела", 6); got != "привет..." {
		t.Errorf("truncate long = %q", got)
	}
}

func TestClampTimeout(t *testing.T) {
	t.Parallel()
	svc := New(DefaultConfig(""), nil, nil, nil, nil, nil)
	cases := []struct {
		in   time.Duration
		want time.Duration
	}{
		{-1, svc.defaultTimeout},
		{0, svc.defaultTimeout},
		{5 * time.Minute, 5 * time.Minute},
		{2 * time.Hour, svc.maxTimeout},
	}
	for _, tc := range cases {
		if got := svc.clampTimeout(tc.in); got != tc.want {
			t.Errorf("clampTimeout(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
