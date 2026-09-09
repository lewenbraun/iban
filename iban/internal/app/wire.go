package app

import (
	"os"
	"path/filepath"

	"github.com/lewenbraun/iban/iban/internal/dictation"
	"github.com/lewenbraun/iban/iban/internal/elevenlabs"
	"github.com/lewenbraun/iban/iban/internal/output"
	"github.com/lewenbraun/iban/iban/internal/recorder"
	"github.com/lewenbraun/iban/iban/internal/state"
)

var defaultStateDir = filepath.Join(os.TempDir(), "iban")

func newService() *dictation.Service {
	return dictation.New(dictation.DefaultConfig(defaultStateDir), recorder.New(),
		newTranscriber, output.NewCopier(), output.NewPaster())
}

func newTranscriber() (dictation.Transcriber, error) {
	key, err := elevenlabs.LoadAPIKey()
	if err != nil {
		return nil, err
	}
	return elevenlabs.New(key), nil
}

func status() state.State {
	return newService().Status()
}
