package app

import (
	"github.com/lewenbraun/go-skeleton/eban/internal/dictation"
	"github.com/lewenbraun/go-skeleton/eban/internal/output"
	"github.com/lewenbraun/go-skeleton/eban/internal/recorder"
	"github.com/lewenbraun/go-skeleton/eban/internal/state"
	"github.com/lewenbraun/go-skeleton/eban/internal/transcribe/elevenlabs"
)

const defaultStateDir = "/tmp/eban"

func newService() *dictation.Service {
	return dictation.New(dictation.DefaultConfig(defaultStateDir), recorder.New(),
		newTranscriber, output.NewCopier(), output.NewTyper(), output.NewNotifier())
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
