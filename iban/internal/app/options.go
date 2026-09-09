package app

import (
	"flag"
	"io"
	"time"

	"github.com/lewenbraun/iban/iban/internal/dictation"
	"github.com/lewenbraun/iban/iban/internal/state"
)

type options struct {
	paste   bool
	enter   bool
	lang    string
	timeout time.Duration
}

func (o *options) mode() state.Mode {
	if o.enter {
		return state.ModePasteEnter
	}
	if o.paste {
		return state.ModePaste
	}
	return state.ModeCopy
}

func parseOptions(args []string) (*options, error) {
	defaults := dictation.DefaultConfig("")
	opts := &options{}
	fs := flag.NewFlagSet("iban", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&opts.paste, "paste", false, "paste text into the saved window")
	fs.BoolVar(&opts.enter, "enter", false, "paste text and press Enter")
	fs.StringVar(&opts.lang, "lang", "", "speech language code")
	fs.DurationVar(&opts.timeout, "timeout", defaults.DefaultTimeout, "max recording length")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return opts, nil
}
