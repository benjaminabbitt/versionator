package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/benjaminabbitt/versionator/pkg/plugin"
	"github.com/spf13/cobra"
)

var (
	patchDryRun             bool
	patchPrereleaseTemplate string
	patchMetadataTemplate   string
)

var PatchCmd = &cobra.Command{
	Use:   "patch [file...]",
	Short: "Patch version in manifest files",
	Long: `Update version strings in project manifest files to match the VERSION file.

If no files are specified, patches all recognized manifest files in the current directory.

Manifest files are discovered from registered patch plugins.
Use 'versionator plugin list patch' to see available patch plugins.

Examples:
  # Patch all manifest files in current directory
  versionator out patch

  # Patch specific file
  versionator out patch pyproject.toml

  # Dry run - show what would change
  versionator out patch --dry-run

  # With prerelease
  versionator out patch --prerelease="alpha"`,
	RunE: runPatch,
}

func runPatch(cmd *cobra.Command, args []string) error {
	if PluginLoader == nil {
		return fmt.Errorf("no plugins loaded")
	}

	if len(PluginLoader.PatchPlugins) == 0 {
		return fmt.Errorf("no patch plugins loaded")
	}

	// Load version data
	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	// Build version string
	versionStr := vd.CoreVersion()

	// Handle prerelease
	if cmd.Flags().Changed("prerelease") {
		if patchPrereleaseTemplate != "" && patchPrereleaseTemplate != UseDefaultMarker {
			versionStr += "-" + patchPrereleaseTemplate
		}
	}

	// Handle metadata
	if cmd.Flags().Changed("metadata") {
		if patchMetadataTemplate != "" && patchMetadataTemplate != UseDefaultMarker {
			versionStr += "+" + patchMetadataTemplate
		}
	}

	// Find files to patch
	var filesToPatch []string
	if len(args) > 0 {
		filesToPatch = args
	} else {
		filesToPatch, err = findManifestFiles(".")
		if err != nil {
			return err
		}
	}

	if len(filesToPatch) == 0 {
		return fmt.Errorf("no manifest files found")
	}

	// Patch each file
	patchedCount := 0
	var lastErr error
	for _, file := range filesToPatch {
		patched, err := patchFile(file, versionStr, patchDryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error patching %s: %v\n", file, err)
			lastErr = err
			continue
		}
		if patched {
			patchedCount++
		}
	}

	// If any file had an error, return the last error
	if lastErr != nil {
		return lastErr
	}

	if patchedCount == 0 {
		fmt.Println("No files were patched (versions may already be up to date)")
	} else if patchDryRun {
		fmt.Printf("Would patch %d file(s)\n", patchedCount)
	} else {
		fmt.Printf("Patched %d file(s) to version %s\n", patchedCount, versionStr)
	}

	return nil
}

func findManifestFiles(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		// Check against known patch plugin file patterns
		for pattern := range PluginLoader.PatchPlugins {
			if matchesFilename(name, pattern) {
				files = append(files, filepath.Join(dir, name))
				break
			}
		}
	}

	return files, nil
}

func matchesFilename(name, pattern string) bool {
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(name, pattern[1:])
	}
	return name == pattern
}

// findPatchPlugin finds the patch plugin that matches the given filename
func findPatchPlugin(filename string) plugin.PatchPluginInterface {
	for pattern, patchPlugin := range PluginLoader.PatchPlugins {
		if matchesFilename(filename, pattern) {
			return patchPlugin
		}
	}
	return nil
}

func patchFile(path string, version string, dryRun bool) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	filename := filepath.Base(path)
	original := string(content)

	// Find matching patch plugin
	patchPlugin := findPatchPlugin(filename)
	if patchPlugin == nil {
		return false, fmt.Errorf("no patch plugin found for %s", filename)
	}

	// Use the plugin's patch function
	patched, err := patchPlugin.Patch(original, version)
	if err != nil {
		return false, err
	}

	if patched == "" || patched == original {
		return false, nil
	}

	if dryRun {
		fmt.Printf("%s: %s\n", path, version)
		return true, nil
	}

	// Write patched content
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, []byte(patched), info.Mode()); err != nil {
		return false, err
	}

	fmt.Printf("Patched %s\n", path)
	return true, nil
}

func init() {
	PatchCmd.Flags().BoolVar(&patchDryRun, "dry-run", false, "Show what would be patched without making changes")

	PatchCmd.Flags().StringVar(&patchPrereleaseTemplate, "prerelease", "", "Pre-release suffix to add")
	PatchCmd.Flag("prerelease").NoOptDefVal = UseDefaultMarker

	PatchCmd.Flags().StringVar(&patchMetadataTemplate, "metadata", "", "Metadata suffix to add")
	PatchCmd.Flag("metadata").NoOptDefVal = UseDefaultMarker
}
