package app

import (
	"flag"
	"io"
	"time"

	"github.com/lewenbraun/go-skeleton/eban/internal/dictation"
	"github.com/lewenbraun/go-skeleton/eban/internal/state"
)

type options struct {
	paste   bool
	lang    string
	timeout time.Duration
}

func (o *options) mode() state.Mode {
	if o.paste {
		return state.ModePaste
	}
	return state.ModeCopy
}

func parseOptions(args []string) (*options, error) {
	defaults := dictation.DefaultConfig("")
	opts := &options{}
	fs := flag.NewFlagSet("eban", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&opts.paste, "paste", false, "type text into focused window")
	fs.StringVar(&opts.lang, "lang", "", "speech language code")
	fs.DurationVar(&opts.timeout, "timeout", defaults.DefaultTimeout, "max recording length")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return opts, nil
}
