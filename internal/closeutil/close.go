package closeutil

import (
	"io"
	"log"
)

// CloseWithLog closes a resource and logs failures with a package prefix.
func CloseWithLog(prefix string, closer io.Closer, resource string) {
	if err := closer.Close(); err != nil {
		log.Printf("%s: failed to close %s: %v", prefix, resource, err)
	}
}
