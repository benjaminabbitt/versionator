package ci

import (
	"fmt"
	"io"
	"os"
)

// envFileMode is the permission used when a CI environment file has to be
// created; CI runners normally pre-create these files, so this applies only to
// the first write.
const envFileMode = 0644

// writeCloser is the slice of *os.File this package needs, named so the close
// path can be exercised without a real filesystem fault.
type writeCloser interface {
	io.Writer
	io.Closer
}

// fileOpener opens a CI environment file for appending.
type fileOpener func(path string) (writeCloser, error)

// appendToFile appends whatever write produces to the CI environment file at
// path, reporting failures against label (e.g. "GITHUB_ENV").
//
// The close error is not discarded, and that is the point of this helper: a
// write is not durable until Close reports success. The kernel can defer a
// write failure — a full disk or an exceeded quota — until close(2), so
// dropping that error lets versionator report success over a truncated
// environment file, and the variables silently never reach the job.
func appendToFile(path, label string, write func(io.Writer) error) error {
	return appendToFileWithOpener(path, label, openAppend, write)
}

func openAppend(path string) (writeCloser, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, envFileMode)
}

func appendToFileWithOpener(path, label string, open fileOpener, write func(io.Writer) error) (err error) {
	file, openErr := open(path)
	if openErr != nil {
		return fmt.Errorf("failed to open %s: %w", label, openErr)
	}

	// A write error is the earlier and more specific failure, so it wins; the
	// close error is reported only when the write itself succeeded.
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close %s: %w", label, closeErr)
		}
	}()

	return write(file)
}
