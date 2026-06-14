package storageutil

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
)

type testCloser struct {
	err    error
	closed bool
}

func (c *testCloser) Close() error {
	c.closed = true
	return c.err
}

func TestCloseWithLogSuccess(t *testing.T) {
	var buf bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	})

	closer := &testCloser{}
	CloseWithLog(closer, "resource")

	if !closer.closed {
		t.Fatal("expected closer to be closed")
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no log output, got: %q", buf.String())
	}
}

func TestCloseWithLogError(t *testing.T) {
	var buf bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	})

	closer := &testCloser{err: errors.New("close failed")}
	CloseWithLog(closer, "resource")

	if !closer.closed {
		t.Fatal("expected closer to be closed")
	}

	got := buf.String()
	if !strings.Contains(got, "storageutil") || !strings.Contains(got, "resource") || !strings.Contains(got, "close failed") {
		t.Fatalf("unexpected log output: %q", got)
	}
}

func TestCloseWithLogNilCloser(t *testing.T) {
	CloseWithLog(nil, "resource")
}
