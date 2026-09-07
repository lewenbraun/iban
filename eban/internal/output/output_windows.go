//go:build windows

// Package output delivers text to the user through clipboard and target-aware
// paste backends built on win32 APIs.
package output

// NewCopier returns the clipboard backend for Windows.
func NewCopier() Copier {
	return winClipboard{}
}

// NewPaster returns the target-aware paste backend for Windows.
func NewPaster() Paster {
	return winPaster{}
}
