package closeutil

import (
	"io"
	"log"
)

// WithLog closes a resource and logs any close error with package context.
func WithLog(packageName string, closer io.Closer, resource string) {
	if closer == nil {
		return
	}
	if err := closer.Close(); err != nil {
		log.Printf("%s: failed to close %s: %v", packageName, resource, err)
	}
}
