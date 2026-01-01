// Package testing provides test utilities for filesystem operations
package testing

import (
	"io/fs"

	"github.com/benjaminabbitt/versionator/internal/filesystem"
	"github.com/spf13/afero"
)

// MemFs wraps afero.MemMapFs to implement filesystem.Fs
type MemFs struct {
	afero.Fs
}

// NewMemFs creates an in-memory filesystem for testing
func NewMemFs() *MemFs {
	return &MemFs{Fs: afero.NewMemMapFs()}
}

func (m *MemFs) ReadFile(filename string) ([]byte, error) {
	return afero.ReadFile(m.Fs, filename)
}

func (m *MemFs) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	return afero.WriteFile(m.Fs, filename, data, perm)
}

// SetupTestFs sets up an in-memory filesystem for testing
// Returns a cleanup function to restore the original filesystem
func SetupTestFs() (*MemFs, func()) {
	memFs := NewMemFs()
	filesystem.SetFs(memFs)
	return memFs, func() {
		filesystem.ResetFs()
	}
}
