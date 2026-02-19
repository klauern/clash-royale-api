package main

import (
	"encoding/csv"
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

func writeCSVDocument(w io.Writer, header []string, rows [][]string) error {
	csvWriter := csv.NewWriter(w)
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for i, row := range rows {
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row %d: %w", i+1, err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return nil
}
