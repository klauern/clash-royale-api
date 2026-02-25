package closeutil

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
	log.SetOutput(&buf)
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
	})

	closer := &testCloser{}
	CloseWithLog("closeutil", closer, "resource")

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
	log.SetOutput(&buf)
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
	})

	closer := &testCloser{err: errors.New("close failed")}
	CloseWithLog("closeutil", closer, "resource")

	if !closer.closed {
		t.Fatal("expected closer to be closed")
	}

	got := buf.String()
	if !strings.Contains(got, "closeutil") || !strings.Contains(got, "resource") || !strings.Contains(got, "close failed") {
		t.Fatalf("unexpected log output: %q", got)
	}
}

func TestCloseWithLogNilCloser(t *testing.T) {
	CloseWithLog("closeutil", nil, "resource")
}
