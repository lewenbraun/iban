// Command eban is a push-to-talk dictation tool backed by the ElevenLabs
// Scribe v2 speech-to-text API.
package main

import (
	"os"

	"github.com/lewenbraun/go-skeleton/eban/internal/app"
)

func main() {
	os.Exit(app.Main(os.Args[1:]))
}
