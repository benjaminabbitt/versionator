package update

import (
	"fmt"

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
	configs       []config.UpdateConfig
	parser        *DaselFileParser
	logger        *zap.Logger
	updatedFiles  []string
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

// UpdateFiles applies all configured updates using the given template data
func (u *Updater) UpdateFiles(data emit.TemplateData) error {
	u.updatedFiles = make([]string, 0)

	for i, cfg := range u.configs {
		if err := u.updateSingleFile(cfg, data); err != nil {
			return fmt.Errorf("updates[%d] (%s): %w", i, cfg.File, err)
		}
		u.updatedFiles = append(u.updatedFiles, cfg.File)
		u.logger.Info(LogFileUpdated,
			zap.String("file", cfg.File),
			zap.String("path", cfg.Path),
		)
	}

	u.logger.Info(LogUpdatesApplied, zap.Int("count", len(u.configs)))
	return nil
}

// updateSingleFile applies a single update configuration
func (u *Updater) updateSingleFile(cfg config.UpdateConfig, data emit.TemplateData) error {
	// Render the template to get the new value
	newValue, err := emit.RenderTemplateWithData(cfg.Template, data)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrTemplateRender, err)
	}

	// Read the file
	var fileData any
	var format Format
	if cfg.Format != "" {
		fileData, format, err = u.parser.ReadWithFormat(cfg.File, cfg.Format)
	} else {
		fileData, format, err = u.parser.Read(cfg.File)
	}
	if err != nil {
		return err
	}

	// Convert to map for modification
	dataMap, ok := fileData.(map[string]any)
	if !ok {
		return fmt.Errorf("file content is not a map structure")
	}

	// Update the value at the specified path
	if err := u.parser.Put(&dataMap, cfg.Path, newValue); err != nil {
		return err
	}

	// Write the file back
	if err := u.parser.Write(cfg.File, dataMap, format); err != nil {
		return err
	}

	return nil
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
