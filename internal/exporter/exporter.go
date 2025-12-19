package exporter

import (
	"fmt"
	"sync"
)

// Exporter defines the interface for all data exporters
type Exporter interface {
	// Export exports the given data to the specified directory
	Export(dataDir string, data interface{}) error

	// Headers returns the column headers for the export format
	Headers() []string

	// Filename returns the default filename for the exported data
	Filename() string
}

// Manager manages multiple export formats and provides a unified interface
type Manager struct {
	formats map[string]Exporter
	mu      sync.RWMutex
}

// NewManager creates a new export manager
func NewManager() *Manager {
	return &Manager{
		formats: make(map[string]Exporter),
	}
}

// Register registers a new exporter for the given format
func (m *Manager) Register(format string, exporter Exporter) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.formats[format]; exists {
		panic(fmt.Sprintf("exporter for format '%s' already registered", format))
	}

	m.formats[format] = exporter
}

// Get returns the exporter for the given format
func (m *Manager) Get(format string) (Exporter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exporter, exists := m.formats[format]
	if !exists {
		return nil, fmt.Errorf("no exporter registered for format '%s'", format)
	}

	return exporter, nil
}

// Export exports data using the specified format
func (m *Manager) Export(format string, dataDir string, data interface{}) error {
	exporter, err := m.Get(format)
	if err != nil {
		return err
	}

	return exporter.Export(dataDir, data)
}

// ListFormats returns all registered export formats
func (m *Manager) ListFormats() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	formats := make([]string, 0, len(m.formats))
	for format := range m.formats {
		formats = append(formats, format)
	}

	return formats
}
