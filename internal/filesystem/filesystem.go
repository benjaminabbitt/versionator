package filesystem

import (
	"io/fs"
	"os"
)

// Fs abstracts filesystem operations for testability
// This interface is compatible with afero.Fs for testing
type Fs interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm fs.FileMode) error
	Stat(name string) (fs.FileInfo, error)
	MkdirAll(path string, perm fs.FileMode) error
}

// osFs implements Fs using the real OS filesystem
type osFs struct{}

func (osFs) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (osFs) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (osFs) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (osFs) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

// AppFs is the filesystem used by the application
// Replace with mock in tests using SetFs()
var AppFs Fs = osFs{}

// SetFs sets the filesystem (for testing)
func SetFs(fs Fs) {
	AppFs = fs
}

// ResetFs resets the filesystem to OS
func ResetFs() {
	AppFs = osFs{}
}
