package recorder

import (
	"os"
	"testing"
)

func TestAlive(t *testing.T) {
	t.Parallel()
	if !Alive(os.Getpid()) {
		t.Error("own pid should be alive")
	}
	if Alive(1 << 21) {
		t.Error("unallocated pid should not be alive")
	}
}
