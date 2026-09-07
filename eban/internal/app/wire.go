package app

import (
	"os"
	"path/filepath"

	"github.com/lewenbraun/eban/eban/internal/dictation"
	"github.com/lewenbraun/eban/eban/internal/elevenlabs"
	"github.com/lewenbraun/eban/eban/internal/output"
	"github.com/lewenbraun/eban/eban/internal/recorder"
	"github.com/lewenbraun/eban/eban/internal/state"
)

var defaultStateDir = filepath.Join(os.TempDir(), "eban")

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
