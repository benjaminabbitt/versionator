package ci

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendToFile_NewPath_CreatesFileWithContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env")

	err := appendToFile(path, "TEST_ENV", func(w io.Writer) error {
		_, err := io.WriteString(w, "FOO=bar\n")
		return err
	})
	if err != nil {
		t.Fatalf("appendToFile returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back file: %v", err)
	}
	if got := string(content); got != "FOO=bar\n" {
		t.Errorf("content = %q, want %q", got, "FOO=bar\n")
	}
}

func TestAppendToFile_ExistingFile_AppendsRatherThanTruncates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env")
	if err := os.WriteFile(path, []byte("EXISTING=1\n"), 0644); err != nil {
		t.Fatalf("seeding file: %v", err)
	}

	err := appendToFile(path, "TEST_ENV", func(w io.Writer) error {
		_, err := io.WriteString(w, "ADDED=2\n")
		return err
	})
	if err != nil {
		t.Fatalf("appendToFile returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back file: %v", err)
	}
	want := "EXISTING=1\nADDED=2\n"
	if got := string(content); got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestAppendToFile_UnopenablePath_ReturnsLabelledError(t *testing.T) {
	// A directory cannot be opened for writing.
	path := t.TempDir()

	err := appendToFile(path, "TEST_ENV", func(io.Writer) error {
		t.Error("write callback must not run when the file cannot be opened")
		return nil
	})
	if err == nil {
		t.Fatal("expected an error opening a directory for writing, got nil")
	}
	if !strings.Contains(err.Error(), "TEST_ENV") {
		t.Errorf("error %q does not name the label TEST_ENV", err)
	}
}

func TestAppendToFile_WriteFails_PropagatesWriteError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env")
	sentinel := errors.New("write blew up")

	err := appendToFile(path, "TEST_ENV", func(io.Writer) error {
		return fmt.Errorf("failed to write to TEST_ENV: %w", sentinel)
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("error = %v, want it to wrap %v", err, sentinel)
	}
}

// The close error is what this helper exists to stop discarding: a write is not
// durable until Close reports success, so a deferred failure (full disk,
// exceeded quota) surfaces there and nowhere else.
func TestAppendToFile_CloseFails_ReturnsCloseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env")
	sentinel := errors.New("close blew up")

	err := appendToFileWithOpener(path, "TEST_ENV",
		func(string) (writeCloser, error) {
			return failingCloser{err: sentinel}, nil
		},
		func(w io.Writer) error {
			_, err := io.WriteString(w, "FOO=bar\n")
			return err
		})

	if !errors.Is(err, sentinel) {
		t.Errorf("error = %v, want it to wrap %v", err, sentinel)
	}
}

// A write error must win over a close error: it is the earlier and more
// specific failure.
func TestAppendToFile_WriteAndCloseBothFail_ReportsWriteError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env")
	writeErr := errors.New("write blew up")
	closeErr := errors.New("close blew up")

	err := appendToFileWithOpener(path, "TEST_ENV",
		func(string) (writeCloser, error) {
			return failingCloser{err: closeErr}, nil
		},
		func(io.Writer) error { return writeErr })

	if !errors.Is(err, writeErr) {
		t.Errorf("error = %v, want it to wrap the write error %v", err, writeErr)
	}
}

type failingCloser struct {
	err error
}

func (f failingCloser) Write(p []byte) (int, error) { return len(p), nil }
func (f failingCloser) Close() error                { return f.err }
