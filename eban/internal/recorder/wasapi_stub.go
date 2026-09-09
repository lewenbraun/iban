//go:build !windows

package recorder

// DefaultCaptureEndpointIDs has no implementation outside Windows.
func DefaultCaptureEndpointIDs() []string {
	return nil
}
