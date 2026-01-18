package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func printf(format string, args ...any) {
	if _, err := fmt.Fprintf(os.Stdout, format, args...); err != nil {
		log.Printf("stdout write failed: %v", err)
	}
}

func fprintf(w io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		log.Printf("write failed: %v", err)
	}
}

func fprintln(w io.Writer, args ...any) {
	if _, err := fmt.Fprintln(w, args...); err != nil {
		log.Printf("write failed: %v", err)
	}
}

type flusher interface {
	Flush() error
}

func flushWriter(w flusher) {
	if err := w.Flush(); err != nil {
		log.Printf("flush failed: %v", err)
	}
}

type csvFlusher interface {
	Flush()
}

func flushCSVWriter(w csvFlusher) {
	w.Flush()
}

func closeFile(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("close failed: %v", err)
	}
}

func setEnv(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		log.Printf("failed to set %s: %v", key, err)
	}
}
