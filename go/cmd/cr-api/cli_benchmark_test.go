//go:build integration

package main

import (
	"context"
	"os"
	"testing"
	"time"
)

// BenchmarkCLICommands benchmarks CLI command performance
func BenchmarkCLICommands(b *testing.B) {
	tempDir := b.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
	os.Setenv("DATA_DIR", tempDir)

	cmd := createTestCommand(tempDir)

	benchmarks := []struct {
		name string
		args []string
	}{
		{
			name: "PlayerCommand",
			args: []string{
				"player",
				"--tag", "BENCH123",
				"--export-csv",
			},
		},
		{
			name: "CardsCommand",
			args: []string{
				"cards",
				"--export-csv",
			},
		},
		{
			name: "AnalyzeCommand",
			args: []string{
				"analyze",
				"--tag", "BENCH123",
				"--export-csv",
			},
		},
		{
			name: "DeckBuildCommand",
			args: []string{
				"deck",
				"build",
				"--tag", "BENCH123",
				"--strategy", "balanced",
				"--export-csv",
			},
		},
		{
			name: "ExportAllCommand",
			args: []string{
				"export",
				"all",
				"--tag", "BENCH123",
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := cmd.Run(context.Background(), append([]string{"cr-api"}, bm.args...))
				if err != nil {
					b.Fatalf("Command failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkCSVExport benchmarks CSV export performance
func BenchmarkCSVExport(b *testing.B) {
	tempDir := b.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
	os.Setenv("DATA_DIR", tempDir)

	cmd := createTestCommand(tempDir)

	// Benchmark different export sizes
	sizes := []struct {
		name string
		args []string
	}{
		{
			name: "SmallExport",
			args: []string{
				"player",
				"--tag", "SMALL123",
				"--export-csv",
			},
		},
		{
			name: "MediumExport",
			args: []string{
				"export",
				"all",
				"--tag", "MED123",
			},
		},
		{
			name: "LargeExport",
			args: []string{
				"export",
				"battles",
				"--tag", "LARGE123",
				"--limit", "1000",
			},
		},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := cmd.Run(context.Background(), append([]string{"cr-api"}, size.args...))
				if err != nil {
					b.Fatalf("Export failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkConcurrentCommands benchmarks concurrent command execution
func BenchmarkConcurrentCommands(b *testing.B) {
	tempDir := b.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
	os.Setenv("DATA_DIR", tempDir)

	cmd := createTestCommand(tempDir)

	b.Run("ConcurrentPlayerCommands", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				err := cmd.Run(context.Background(), []string{
					"cr-api",
					"player",
					"--tag", "CONCURRENT123",
					"--export-csv",
				})
				if err != nil {
					b.Fatalf("Concurrent command failed: %v", err)
				}
			}
		})
	})

	b.Run("ConcurrentMixedCommands", func(b *testing.B) {
		commands := [][]string{
			{"cr-api", "player", "--tag", "MIX123", "--export-csv"},
			{"cr-api", "cards", "--export-csv"},
			{"cr-api", "analyze", "--tag", "MIX123", "--export-csv"},
		}

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				args := commands[i%len(commands)]
				err := cmd.Run(context.Background(), args)
				if err != nil {
					b.Fatalf("Concurrent command failed: %v", err)
				}
				i++
			}
		})
	})
}

// BenchmarkMemoryUsage benchmarks memory usage for different operations
func BenchmarkMemoryUsage(b *testing.B) {
	tempDir := b.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
	os.Setenv("DATA_DIR", tempDir)

	cmd := createTestCommand(tempDir)

	b.Run("PlayerExportMemory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := cmd.Run(context.Background(), []string{
				"cr-api",
				"player",
				"--tag", "MEM123",
				"--export-csv",
				"--save",
			})
			if err != nil {
				b.Fatalf("Command failed: %v", err)
			}
		}
	})

	b.Run("ExportAllMemory", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			err := cmd.Run(context.Background(), []string{
				"cr-api",
				"export",
				"all",
				"--tag", "MEMALL123",
			})
			if err != nil {
				b.Fatalf("Command failed: %v", err)
			}
		}
	})
}

// BenchmarkStartupTime benchmarks CLI startup time
func BenchmarkStartupTime(b *testing.B) {
	tempDir := b.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")

	for i := 0; i < b.N; i++ {
		start := time.Now()
		_ = createTestCommand(tempDir)
		duration := time.Since(start)

		// Report startup time
		b.ReportMetric(float64(duration.Nanoseconds()), "ns/op")
	}
}

// TestCLIProfiling runs CPU and memory profiling on CLI operations
func TestCLIProfiling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping profiling test in short mode")
	}

	tempDir := t.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
	os.Setenv("DATA_DIR", tempDir)

	cmd := createTestCommand(tempDir)

	// Profile CPU usage
	t.Run("CPUProfile", func(t *testing.T) {
		// In a real scenario, you would enable CPU profiling here
		// For testing purposes, we'll just run the command multiple times

		for i := 0; i < 10; i++ {
			err := cmd.Run(context.Background(), []string{
				"cr-api",
				"export",
				"all",
				"--tag", "CPU123",
			})
			if err != nil {
				t.Errorf("Command failed during profiling: %v", err)
			}
		}
	})

	// Profile memory usage
	t.Run("MemoryProfile", func(t *testing.T) {
		// In a real scenario, you would enable memory profiling here
		// For testing purposes, we'll run memory-intensive operations

		for i := 0; i < 5; i++ {
			err := cmd.Run(context.Background(), []string{
				"cr-api",
				"export",
				"battles",
				"--tag", "MEM123",
				"--limit", "100",
			})
			if err != nil {
				t.Errorf("Command failed during memory profiling: %v", err)
			}
		}
	})
}

// BenchmarkDataProcessing benchmarks data processing performance
func BenchmarkDataProcessing(b *testing.B) {
	tempDir := b.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
	os.Setenv("DATA_DIR", tempDir)

	cmd := createTestCommand(tempDir)

	b.Run("PlayerDataProcessing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := cmd.Run(context.Background(), []string{
				"cr-api",
				"player",
				"--tag", "PROC123",
				"--export-csv",
				"--save",
			})
			if err != nil {
				b.Fatalf("Data processing failed: %v", err)
			}
		}
	})

	b.Run("AnalysisDataProcessing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := cmd.Run(context.Background(), []string{
				"cr-api",
				"analyze",
				"--tag", "ANALYSIS123",
				"--export-csv",
				"--save",
			})
			if err != nil {
				b.Fatalf("Analysis processing failed: %v", err)
			}
		}
	})

	b.Run("EventScanProcessing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := cmd.Run(context.Background(), []string{
				"cr-api",
				"events",
				"scan",
				"--tag", "EVENT123",
				"--days", "30",
				"--export-csv",
			})
			if err != nil {
				b.Fatalf("Event scan processing failed: %v", err)
			}
		}
	})
}
