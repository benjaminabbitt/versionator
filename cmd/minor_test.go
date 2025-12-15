package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinorCommand(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		initialVersion  string
		expectedVersion string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "increment from 1.2.3",
			args:            []string{"minor", "increment"},
			initialVersion:  "1.2.3",
			expectedVersion: "1.3.0",
			expectError:     false,
		},
		{
			name:            "increment with inc alias",
			args:            []string{"minor", "inc"},
			initialVersion:  "0.5.7",
			expectedVersion: "0.6.0",
			expectError:     false,
		},
		{
			name:            "increment with + alias",
			args:            []string{"minor", "+"},
			initialVersion:  "2.1.9",
			expectedVersion: "2.2.0",
			expectError:     false,
		},
		{
			name:            "decrement from 1.3.5",
			args:            []string{"minor", "decrement"},
			initialVersion:  "1.3.5",
			expectedVersion: "1.2.0",
			expectError:     false,
		},
		{
			name:            "decrement with dec alias",
			args:            []string{"minor", "dec"},
			initialVersion:  "2.5.1",
			expectedVersion: "2.4.0",
			expectError:     false,
		},
		{
			name:            "increment from default version",
			args:            []string{"minor", "increment"},
			initialVersion:  "", // No VERSION file
			expectedVersion: "0.1.0",
			expectError:     false,
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
metadata:
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
				err = os.WriteFile("VERSION", []byte(tt.initialVersion+"\n"), 0644)
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
				assert.Equal(t, tt.expectedVersion, strings.TrimSpace(string(content)))
			}

			// Reset command state
			rootCmd.SetOut(nil)
			rootCmd.SetErr(nil)
			rootCmd.SetArgs(nil)
		})
	}
}

func TestMinorCommandHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "minor help",
			args: []string{"minor", "--help"},
		},
		{
			name: "minor increment help",
			args: []string{"minor", "increment", "--help"},
		},
		{
			name: "minor decrement help",
			args: []string{"minor", "decrement", "--help"},
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
