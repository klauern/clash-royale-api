package storageutil

import (
	"io"

	"github.com/klauer/clash-royale-api/go/internal/closeutil"
)

// CloseWithLog closes a resource and logs any close error with storageutil context.
func CloseWithLog(closer io.Closer, resourceName string) {
	closeutil.WithLog("storageutil", closer, resourceName)
}
