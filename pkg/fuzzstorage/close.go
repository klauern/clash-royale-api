package fuzzstorage

import (
	"io"
	"log"
)

func closeWithLog(closer io.Closer, resource string) {
	if err := closer.Close(); err != nil {
		log.Printf("fuzzstorage: failed to close %s: %v", resource, err)
	}
}
