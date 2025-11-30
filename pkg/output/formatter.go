package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/yourusername/subfinder-pro/pkg/runner"
)

// Formatter interface for output formatting
type Formatter interface {
	Format(results []runner.SubdomainResult, writer io.Writer) error
}

// TextFormatter formats output as plain text
type TextFormatter struct {
	sorted bool
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter(sorted bool) *TextFormatter {
	return &TextFormatter{sorted: sorted}
}

// Format formats results as plain text
func (tf *TextFormatter) Format(results []runner.SubdomainResult, writer io.Writer) error {
	// Extract just the hosts
	hosts := make([]string, len(results))
	for i, result := range results {
		hosts[i] = result.Host
	}
	
	// Sort if requested
	if tf.sorted {
		sort.Strings(hosts)
	}
	
	// Write each host on a new line
	for _, host := range hosts {
		if _, err := fmt.Fprintln(writer, host); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	
	return nil
}

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	sorted bool
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(sorted bool) *JSONFormatter {
	return &JSONFormatter{sorted: sorted}
}

// Format formats results as JSONL (JSON Lines)
func (jf *JSONFormatter) Format(results []runner.SubdomainResult, writer io.Writer) error {
	// Sort if requested
	if jf.sorted {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Host < results[j].Host
		})
	}
	
	// Write each result as a JSON line
	encoder := json.NewEncoder(writer)
	for _, result := range results {
		if err := encoder.Encode(result); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	}
	
	return nil
}

// Writer handles output writing to file or stdout
type Writer struct {
	file   *os.File
	writer io.Writer
}

// NewWriter creates a new output writer
func NewWriter(path string) (*Writer, error) {
	if path == "" || path == "-" {
		// Write to stdout
		return &Writer{
			writer: os.Stdout,
		}, nil
	}
	
	// Create directory if needed
	// dir := filepath.Dir(path)
	// if err := os.MkdirAll(dir, 0755); err != nil {
	// 	return nil, fmt.Errorf("failed to create directory: %w", err)
	// }
	
	// Open file for writing
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	
	return &Writer{
		file:   file,
		writer: file,
	}, nil
}

// Write writes data to the output
func (w *Writer) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}

// Close closes the output writer
func (w *Writer) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// WriteResults writes formatted results to output
func WriteResults(results []runner.SubdomainResult, formatter Formatter, outputPath string) error {
	writer, err := NewWriter(outputPath)
	if err != nil {
		return err
	}
	defer writer.Close()
	
	return formatter.Format(results, writer)
}

// WriteSimple writes simple string results to output
func WriteSimple(subdomains []string, outputPath string, sorted bool) error {
	writer, err := NewWriter(outputPath)
	if err != nil {
		return err
	}
	defer writer.Close()
	
	if sorted {
		sort.Strings(subdomains)
	}
	
	for _, subdomain := range subdomains {
		if _, err := fmt.Fprintln(writer, subdomain); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	
	return nil
}
