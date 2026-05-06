package storageutil

import (
	"io"
	"log"
)

// CloseWithLog closes a resource and logs any close error as a warning.
func CloseWithLog(closer io.Closer, resourceName string) {
	if err := closer.Close(); err != nil {
		log.Printf("warning: failed to close %s: %v", resourceName, err)
	}
}
