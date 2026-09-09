package output

// Copier places text into the system clipboard.
type Copier interface {
	Copy(text string) error
}

// Paster captures a destination window and pastes into it later.
type Paster interface {
	CaptureTarget() (string, error)
	PasteTarget(target string, pressEnter bool) error
}
