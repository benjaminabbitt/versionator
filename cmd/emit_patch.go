package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/benjaminabbitt/versionator/pkg/plugin"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/spf13/cobra"
)

var (
	patchDryRun             bool
	patchPrereleaseTemplate string
	patchMetadataTemplate   string
)

var emitPatchCmd = &cobra.Command{
	Use:   "patch [file...]",
	Short: "Patch version in manifest files",
	Long: `Update version strings in project manifest files to match the VERSION file.

If no files are specified, patches all recognized manifest files in the current directory.

Manifest files are discovered from registered language plugins.

Examples:
  # Patch all manifest files in current directory
  versionator emit patch

  # Patch specific file
  versionator emit patch pyproject.toml

  # Dry run - show what would change
  versionator emit patch --dry-run

  # With prerelease
  versionator emit patch --prerelease="alpha"`,
	RunE: runEmitPatch,
}

func runEmitPatch(cmd *cobra.Command, args []string) error {
	// Load version data
	vd, err := version.Load()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}

	// Build version string
	versionStr := vd.CoreVersion()

	// Handle prerelease
	if cmd.Flags().Changed("prerelease") {
		if patchPrereleaseTemplate != "" && patchPrereleaseTemplate != useDefaultMarker {
			versionStr += "-" + patchPrereleaseTemplate
		}
	}

	// Handle metadata
	if cmd.Flags().Changed("metadata") {
		if patchMetadataTemplate != "" && patchMetadataTemplate != useDefaultMarker {
			versionStr += "+" + patchMetadataTemplate
		}
	}

	// Collect all patch configs from plugins
	patchConfigs := collectPatchConfigs()

	// Find files to patch
	var filesToPatch []string
	if len(args) > 0 {
		filesToPatch = args
	} else {
		filesToPatch, err = findManifestFiles(".", patchConfigs)
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
		patched, err := patchFile(file, versionStr, patchConfigs, patchDryRun)
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

// collectPatchConfigs gathers all patch configurations from registered language plugins
func collectPatchConfigs() []plugin.PatchConfig {
	var configs []plugin.PatchConfig
	for _, lp := range plugin.GetLanguagePlugins() {
		patchConfigs := lp.GetPatchConfigs()
		if patchConfigs != nil {
			configs = append(configs, patchConfigs...)
		}
	}
	return configs
}

func findManifestFiles(dir string, patchConfigs []plugin.PatchConfig) ([]string, error) {
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

		// Check against known patch configs
		for _, pc := range patchConfigs {
			if matchesFilename(name, pc.FilePath) {
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

func patchFile(path string, version string, patchConfigs []plugin.PatchConfig, dryRun bool) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	filename := filepath.Base(path)
	original := string(content)

	// Find matching patch config
	var matchedConfig *plugin.PatchConfig
	for i := range patchConfigs {
		if matchesFilename(filename, patchConfigs[i].FilePath) {
			matchedConfig = &patchConfigs[i]
			break
		}
	}

	if matchedConfig == nil {
		return false, fmt.Errorf("no patch config found for %s", filename)
	}

	// Use the plugin's patch function
	if matchedConfig.Patch == nil {
		return false, fmt.Errorf("no patch function defined for %s", matchedConfig.Name)
	}

	patched, err := matchedConfig.Patch(original, version)
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
	emitCmd.AddCommand(emitPatchCmd)

	emitPatchCmd.Flags().BoolVar(&patchDryRun, "dry-run", false, "Show what would be patched without making changes")

	emitPatchCmd.Flags().StringVar(&patchPrereleaseTemplate, "prerelease", "", "Pre-release suffix to add")
	emitPatchCmd.Flag("prerelease").NoOptDefVal = useDefaultMarker

	emitPatchCmd.Flags().StringVar(&patchMetadataTemplate, "metadata", "", "Metadata suffix to add")
	emitPatchCmd.Flag("metadata").NoOptDefVal = useDefaultMarker
}
