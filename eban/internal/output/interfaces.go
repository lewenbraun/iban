package output

// Copier places text into the system clipboard.
type Copier interface {
	Copy(text string) error
}

// Typer types text into the currently focused window.
type Typer interface {
	Type(text string) error
}

// Notifier shows a desktop notification.
type Notifier interface {
	Notify(body, urgency string)
}
