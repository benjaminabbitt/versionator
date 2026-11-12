package update

import (
	"bytes"
	"fmt"
	"os"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"go.uber.org/zap"
)

// FileUpdater defines the interface for file update operations
type FileUpdater interface {
	// UpdateFiles applies all configured updates using the given template data
	UpdateFiles(data emit.TemplateData) error
	// ValidateConfig checks all update configurations are valid
	ValidateConfig() error
	// GetFilesToCommit returns list of files that were modified
	GetFilesToCommit() []string
}

// Updater implements FileUpdater with actual file operations
type Updater struct {
	configs      []config.UpdateConfig
	parser       *DaselFileParser
	logger       *zap.Logger
	updatedFiles []string
}

// NewUpdater creates an Updater (IoC constructor accepting dependencies)
func NewUpdater(configs []config.UpdateConfig, parser *DaselFileParser, logger *zap.Logger) *Updater {
	return &Updater{
		configs:      configs,
		parser:       parser,
		logger:       logger,
		updatedFiles: make([]string, 0),
	}
}

// NewUpdaterDefault creates an Updater with default dependencies
// This is excluded from unit test coverage as it's tested via integration tests
func NewUpdaterDefault(configs []config.UpdateConfig) (*Updater, error) { //nolint:unused
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return NewUpdater(configs, NewDaselFileParser(), logger), nil
}

// UpdateFiles applies all configured updates using the given template data.
//
// Every update is rendered, checked and applied in memory first; only when all
// of them succeed are the files written. So a refused update (a missing path, a
// value Cargo would refuse) leaves every configured file as it was. Several
// updates of one file apply in order to the same content.
func (u *Updater) UpdateFiles(data emit.TemplateData) error {
	u.updatedFiles = make([]string, 0)

	contents := make(map[string][]byte)
	original := make(map[string][]byte)
	var order []string
	for i, cfg := range u.configs {
		raw, seen := contents[cfg.File]
		if !seen {
			var err error
			if raw, err = readExisting(cfg.File); err != nil {
				return fmt.Errorf("updates[%d] (%s): %w", i, cfg.File, err)
			}
			original[cfg.File] = raw
			order = append(order, cfg.File)
		}
		updated, err := u.planSingleFile(cfg, raw, data)
		if err != nil {
			return fmt.Errorf("updates[%d] (%s): %w", i, cfg.File, err)
		}
		contents[cfg.File] = updated
	}

	for _, file := range order {
		if bytes.Equal(contents[file], original[file]) {
			continue // already at this version: the file is not rewritten
		}
		if err := os.WriteFile(file, contents[file], FilePermission); err != nil {
			return fmt.Errorf("%s: %s: %w", ErrFileWriteFailed, file, err)
		}
	}

	for _, cfg := range u.configs {
		u.updatedFiles = append(u.updatedFiles, cfg.File)
		u.logger.Info(LogFileUpdated,
			zap.String("file", cfg.File),
			zap.String("path", cfg.Path),
		)
	}

	u.logger.Info(LogUpdatesApplied, zap.Int("count", len(u.configs)))
	return nil
}

// planSingleFile returns raw with one update applied.
func (u *Updater) planSingleFile(cfg config.UpdateConfig, raw []byte, data emit.TemplateData) ([]byte, error) {
	// Render the template to get the new value
	newValue, err := emit.RenderTemplateWithData(cfg.Template, data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrTemplateRender, err)
	}

	format, err := u.parser.detectFormat(cfg.File, cfg.Format)
	if err != nil {
		return nil, err
	}

	switch format {
	case FormatTOML:
		// Cargo refuses a 4-part version and a "v" prefix; they are refused
		// here, not stripped, so what Cargo gets is what the template said.
		if isCargoVersionKey(cfg.File, cfg.Path) && !cargoVersionValid(newValue) {
			return nil, fmt.Errorf("%s = %q: %s", cfg.Path, newValue, ErrCargoVersionInvalid)
		}
		// Targeted text replacement preserves comments and ordering.
		return u.parser.setTOMLString(raw, cfg.Path, newValue)
	case FormatJSON:
		// Surgical value replacement preserves key order and formatting. A
		// parse-modify-serialize round-trip would reorder keys (Go marshals maps
		// alphabetically), mangling hand-ordered manifests like package.json.
		return setJSONString(raw, cfg.Path, newValue)
	}

	// YAML: parse-modify-serialize.
	parsed, err := u.parser.unmarshal(raw, format)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", ErrFileParseFailed, cfg.File, err)
	}
	if err := u.parser.Put(&parsed, cfg.Path, newValue); err != nil {
		return nil, err
	}
	out, err := u.parser.marshal(parsed, format)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFileWriteFailed, err)
	}
	return out, nil
}

// ValidateConfig checks all update configurations are valid
func (u *Updater) ValidateConfig() error {
	u.logger.Debug(LogValidatingConfig, zap.Int("count", len(u.configs)))

	for i, cfg := range u.configs {
		// Check file exists
		_, _, err := u.parser.Read(cfg.File)
		if err != nil {
			return fmt.Errorf("updates[%d]: %w", i, err)
		}

		// Note: We don't validate paths here as the file structure may change
		// Path validation happens during actual update
	}

	return nil
}

// GetFilesToCommit returns list of files that were modified
func (u *Updater) GetFilesToCommit() []string {
	return u.updatedFiles
}
