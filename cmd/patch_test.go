package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		initialVersion string
		expectedVersion string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "increment from 1.2.3",
			args:           []string{"patch", "increment"},
			initialVersion: "1.2.3",
			expectedVersion: "1.2.4",
			expectError:    false,
		},
		{
			name:           "increment with inc alias",
			args:           []string{"patch", "inc"},
			initialVersion: "0.5.7",
			expectedVersion: "0.5.8",
			expectError:    false,
		},
		{
			name:           "increment with + alias",
			args:           []string{"patch", "+"},
			initialVersion: "2.1.9",
			expectedVersion: "2.1.10",
			expectError:    false,
		},
		{
			name:           "decrement from 1.3.5",
			args:           []string{"patch", "decrement"},
			initialVersion: "1.3.5",
			expectedVersion: "1.3.4",
			expectError:    false,
		},
		{
			name:           "decrement with dec alias",
			args:           []string{"patch", "dec"},
			initialVersion: "2.5.1",
			expectedVersion: "2.5.0",
			expectError:    false,
		},
		{
			name:           "decrement with - alias",
			args:           []string{"patch", "-"},
			initialVersion: "3.7.2",
			expectedVersion: "3.7.1",
			expectError:    false,
		},
		{
			name:           "increment from default version",
			args:           []string{"patch", "increment"},
			initialVersion: "", // No VERSION file
			expectedVersion: "0.0.1",
			expectError:    false,
		},
		{
			name:           "increment large patch number",
			args:           []string{"patch", "increment"},
			initialVersion: "1.0.999",
			expectedVersion: "1.0.1000",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				os.Chdir(originalDir)
			}()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create config file
			configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
			err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
			require.NoError(t, err)

			// Create VERSION file if initial version is provided
			if tt.initialVersion != "" {
				err = os.WriteFile("VERSION", []byte(tt.initialVersion), 0644)
				require.NoError(t, err)
			}

			// Capture output
			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify VERSION file content
				content, err := os.ReadFile("VERSION")
				require.NoError(t, err)
				actualVersion := strings.TrimSpace(string(content))
				assert.Equal(t, tt.expectedVersion, actualVersion)

				// The main behavior we care about is that the VERSION file was updated correctly
				// Output message testing is less important and more brittle
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetErr(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestPatchCommandHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "patch help",
			args: []string{"patch", "--help"},
		},
		{
			name: "patch increment help",
			args: []string{"patch", "increment", "--help"},
		},
		{
			name: "patch decrement help",
			args: []string{"patch", "decrement", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			assert.Contains(t, output, "Usage:")

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestPatchCommandEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		initialVersion string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "increment with whitespace in version file",
			args:           []string{"patch", "increment"},
			initialVersion: "  1.2.3  \n",
			expectError:    false,
		},
		{
			name:           "decrement with trailing newline",
			args:           []string{"patch", "decrement"},
			initialVersion: "1.2.3\n",
			expectError:    false,
		},
		{
			name:           "invalid command",
			args:           []string{"patch", "invalid"},
			initialVersion: "1.2.3",
			expectError:    true,
			errorContains:  "unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			tempDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				os.Chdir(originalDir)
			}()
			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create config file
			configContent := `prefix: ""
suffix:
  enabled: false
  type: "git"
  git:
    hashLength: 7
logging:
  output: "console"
`
			err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
			require.NoError(t, err)

			// Create VERSION file
			err = os.WriteFile("VERSION", []byte(tt.initialVersion), 0644)
			require.NoError(t, err)

			// Capture output
			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			// Execute command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetErr(nil)
			rootCmd.SetArgs(nil)
		})
	}
}
