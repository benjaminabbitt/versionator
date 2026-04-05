package update

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/tomwright/dasel/v3"
	"gopkg.in/yaml.v3"
)

// Format represents supported file formats
type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
	FormatTOML Format = "toml"
)

// FileParser provides operations on structured files (JSON, YAML, TOML)
type FileParser interface {
	// Read reads and parses a structured file, returning the data and detected format
	Read(filepath string) (any, Format, error)
	// Write writes data back to file in the specified format
	Write(filepath string, data any, format Format) error
	// Select retrieves a value at the given path using dasel syntax
	Select(data any, path string) (any, error)
	// Put updates a value at the given path using dasel syntax
	Put(data *map[string]any, path string, value any) error
}

// DaselFileParser implements FileParser using dasel for path queries
type DaselFileParser struct{}

// NewDaselFileParser creates a new DaselFileParser
func NewDaselFileParser() *DaselFileParser {
	return &DaselFileParser{}
}

// Read reads and parses a structured file
func (p *DaselFileParser) Read(path string) (any, Format, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("%s: %s", ErrFileNotFound, path)
		}
		return nil, "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	format, err := p.detectFormat(path, "")
	if err != nil {
		return nil, "", err
	}

	data, err := p.unmarshal(content, format)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %s: %w", ErrFileParseFailed, path, err)
	}

	return data, format, nil
}

// ReadWithFormat reads a file using an explicit format
func (p *DaselFileParser) ReadWithFormat(path string, explicitFormat string) (any, Format, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("%s: %s", ErrFileNotFound, path)
		}
		return nil, "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	format, err := p.detectFormat(path, explicitFormat)
	if err != nil {
		return nil, "", err
	}

	data, err := p.unmarshal(content, format)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %s: %w", ErrFileParseFailed, path, err)
	}

	return data, format, nil
}

// Write writes data back to file in the specified format
func (p *DaselFileParser) Write(path string, data any, format Format) error {
	content, err := p.marshal(data, format)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrFileWriteFailed, err)
	}

	if err := os.WriteFile(path, content, FilePermission); err != nil {
		return fmt.Errorf("%s: %w", ErrFileWriteFailed, err)
	}

	return nil
}

// Select retrieves a value at the given path using dasel syntax
func (p *DaselFileParser) Select(data any, path string) (any, error) {
	result, count, err := dasel.Select(context.Background(), data, path)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", ErrPathNotFound, path, err)
	}
	if count == 0 {
		return nil, fmt.Errorf("%s: %s", ErrPathNotFound, path)
	}
	// dasel.Select returns a slice of results; for single-value queries, return the first element
	if results, ok := result.([]any); ok && len(results) == 1 {
		return results[0], nil
	}
	return result, nil
}

// Put updates a value at the given path using dasel syntax
func (p *DaselFileParser) Put(data *map[string]any, path string, value any) error {
	_, err := dasel.Modify(context.Background(), data, path, value)
	if err != nil {
		return fmt.Errorf("%s: %s: %w", ErrInvalidSelector, path, err)
	}
	return nil
}

// detectFormat determines the file format from extension or explicit override
func (p *DaselFileParser) detectFormat(path string, explicitFormat string) (Format, error) {
	if explicitFormat != "" {
		switch strings.ToLower(explicitFormat) {
		case "json":
			return FormatJSON, nil
		case "yaml", "yml":
			return FormatYAML, nil
		case "toml":
			return FormatTOML, nil
		default:
			return "", fmt.Errorf("%s: %s", ErrUnsupportedFormat, explicitFormat)
		}
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return FormatJSON, nil
	case ".yaml", ".yml":
		return FormatYAML, nil
	case ".toml":
		return FormatTOML, nil
	default:
		return "", fmt.Errorf("%s: cannot detect format from extension %s", ErrUnsupportedFormat, ext)
	}
}

// unmarshal parses content based on format
func (p *DaselFileParser) unmarshal(content []byte, format Format) (map[string]any, error) {
	var data map[string]any

	switch format {
	case FormatJSON:
		if err := json.Unmarshal(content, &data); err != nil {
			return nil, err
		}
	case FormatYAML:
		if err := yaml.Unmarshal(content, &data); err != nil {
			return nil, err
		}
	case FormatTOML:
		if err := toml.Unmarshal(content, &data); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnsupportedFormat, format)
	}

	return data, nil
}

// marshal converts data to the specified format
func (p *DaselFileParser) marshal(data any, format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		return json.MarshalIndent(data, "", "  ")
	case FormatYAML:
		return yaml.Marshal(data)
	case FormatTOML:
		return toml.Marshal(data)
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnsupportedFormat, format)
	}
}

// UpdateTOMLValue does a targeted value replacement in a TOML file,
// preserving comments, ordering, and formatting.
// Parses to find the old value at the path, then replaces it in the raw bytes.
func (p *DaselFileParser) UpdateTOMLValue(filePath string, path string, newValue string) error {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	// Parse to navigate to the value at the path
	var data map[string]any
	if err := toml.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("%s: %s: %w", ErrFileParseFailed, filePath, err)
	}

	oldValue, err := p.Select(data, path)
	if err != nil {
		return err
	}

	oldStr := fmt.Sprintf("%v", oldValue)
	if oldStr == newValue {
		return nil
	}

	// Replace the quoted old value with the quoted new value in the raw bytes
	oldQuoted := []byte(`"` + oldStr + `"`)
	newQuoted := []byte(`"` + newValue + `"`)

	result := bytes.Replace(raw, oldQuoted, newQuoted, 1)
	if bytes.Equal(result, raw) {
		// Try single quotes
		oldQuoted = []byte(`'` + oldStr + `'`)
		result = bytes.Replace(raw, oldQuoted, newQuoted, 1)
	}

	if bytes.Equal(result, raw) {
		return fmt.Errorf("could not find value %q to replace in %s", oldStr, filePath)
	}

	return os.WriteFile(filePath, result, FilePermission)
}
