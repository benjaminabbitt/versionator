package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CORE FUNCTIONALITY
// =============================================================================

// TestSortStrings_MixedOrder_SortsAlphabetically validates that the sortStrings
// function correctly sorts string slices in ascending alphabetical order.
//
// Why: This is the primary use case for sortStrings - taking an unsorted slice
// and producing a predictable, alphabetically ordered result. Consistent
// ordering is essential for deterministic output in version metadata and
// configuration display.
//
// What: Given a slice of strings in arbitrary order, when sortStrings is called,
// then the slice is modified in-place to be in ascending alphabetical order.
func TestSortStrings_MixedOrder_SortsAlphabetically(t *testing.T) {
	// Precondition: A slice with strings in random order
	input := []string{"banana", "apple", "cherry"}

	// Action: Sort the slice in place
	sortStrings(input)

	// Expected: Strings are in ascending alphabetical order
	expected := []string{"apple", "banana", "cherry"}
	assert.Equal(t, expected, input)
}

// =============================================================================
// KEY VARIATIONS
// =============================================================================

// TestSortStrings_AlreadySorted_RemainsUnchanged verifies that sortStrings
// handles already-sorted input correctly without corrupting the order.
//
// Why: The function must be idempotent - calling it on sorted data should not
// alter the result. This ensures stability when sortStrings is called multiple
// times on the same data.
//
// What: Given a slice already in sorted order, when sortStrings is called,
// then the slice remains in the same order.
func TestSortStrings_AlreadySorted_RemainsUnchanged(t *testing.T) {
	// Precondition: A slice already in sorted order
	input := []string{"a", "b", "c"}

	// Action: Sort the slice
	sortStrings(input)

	// Expected: Order is unchanged
	expected := []string{"a", "b", "c"}
	assert.Equal(t, expected, input)
}

// TestSortStrings_ReverseOrder_SortsCorrectly verifies that sortStrings
// handles worst-case input (reverse sorted) correctly.
//
// Why: Reverse-sorted input represents a common edge case that exercises the
// full sorting logic. This ensures the algorithm handles maximum displacement
// of elements.
//
// What: Given a slice in reverse alphabetical order, when sortStrings is called,
// then the slice is reordered to ascending alphabetical order.
func TestSortStrings_ReverseOrder_SortsCorrectly(t *testing.T) {
	// Precondition: A slice in reverse order
	input := []string{"c", "b", "a"}

	// Action: Sort the slice
	sortStrings(input)

	// Expected: Strings are in ascending order
	expected := []string{"a", "b", "c"}
	assert.Equal(t, expected, input)
}

// TestSortStrings_WithDuplicates_PreservesDuplicates verifies that duplicate
// values are preserved and grouped together after sorting.
//
// Why: Configuration variables or metadata keys may have duplicate entries in
// some contexts. The sort must preserve all values without deduplication.
//
// What: Given a slice containing duplicate strings, when sortStrings is called,
// then all duplicates are preserved and grouped adjacently in sorted order.
func TestSortStrings_WithDuplicates_PreservesDuplicates(t *testing.T) {
	// Precondition: A slice with duplicate values
	input := []string{"b", "a", "b", "a"}

	// Action: Sort the slice
	sortStrings(input)

	// Expected: Duplicates are preserved and grouped together
	expected := []string{"a", "a", "b", "b"}
	assert.Equal(t, expected, input)
}

// =============================================================================
// EDGE CASES
// =============================================================================

// TestSortStrings_EmptySlice_HandlesGracefully verifies that sortStrings
// handles empty input without panicking or producing unexpected results.
//
// Why: Empty slices are a common boundary condition. The function must handle
// this case gracefully to avoid nil pointer errors or panics in calling code.
//
// What: Given an empty slice, when sortStrings is called, then the slice
// remains empty and no error occurs.
func TestSortStrings_EmptySlice_HandlesGracefully(t *testing.T) {
	// Precondition: An empty slice
	input := []string{}

	// Action: Sort the slice
	sortStrings(input)

	// Expected: Slice remains empty
	expected := []string{}
	assert.Equal(t, expected, input)
}

// TestSortStrings_SingleElement_RemainsUnchanged verifies that a single-element
// slice is handled correctly.
//
// Why: Single-element slices are a boundary condition where sorting is trivial.
// The function must not corrupt or panic on minimal input.
//
// What: Given a slice with exactly one element, when sortStrings is called,
// then the slice remains unchanged with that single element.
func TestSortStrings_SingleElement_RemainsUnchanged(t *testing.T) {
	// Precondition: A slice with one element
	input := []string{"a"}

	// Action: Sort the slice
	sortStrings(input)

	// Expected: Slice is unchanged
	expected := []string{"a"}
	assert.Equal(t, expected, input)
}

// =============================================================================
// MINUTIAE
// =============================================================================

// TestSortStrings_MixedCase_UppercaseBeforeLowercase verifies the lexicographic
// ordering behavior where uppercase letters sort before lowercase.
//
// Why: Go's standard sort uses byte-wise comparison, meaning uppercase ASCII
// characters (65-90) sort before lowercase (97-122). Callers must understand
// this behavior to avoid surprises with mixed-case input.
//
// What: Given a slice with mixed uppercase and lowercase strings, when
// sortStrings is called, then uppercase letters appear before their lowercase
// equivalents due to ASCII ordering.
func TestSortStrings_MixedCase_UppercaseBeforeLowercase(t *testing.T) {
	// Precondition: A slice with mixed case characters
	input := []string{"b", "A", "a", "B"}

	// Action: Sort the slice
	sortStrings(input)

	// Expected: ASCII ordering places uppercase before lowercase
	expected := []string{"A", "B", "a", "b"}
	assert.Equal(t, expected, input)
}

// =============================================================================
// RUNVARS TESTS
// =============================================================================

// TestRunVars_DisplaysTemplateVariables verifies that the vars command outputs
// all available template variables organized by category.
//
// Why: Users need to discover what template variables are available when
// creating custom version output templates.
//
// What: Run "config vars" in a directory with VERSION file, verify output
// contains expected variable categories and names.
func TestRunVars_DisplaysTemplateVariables(t *testing.T) {
	// Precondition: temp directory with VERSION file
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.2.3\n"), 0644)
	require.NoError(t, err)

	// Action: Execute "config vars"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "vars"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and output contains template variables
	assert.NoError(t, err)
	output := stdout.String()

	// Check for expected categories
	assert.Contains(t, output, "Template Variables")
	assert.Contains(t, output, "Version Components")
	assert.Contains(t, output, "VCS/Git")

	// Check for expected variable names
	assert.Contains(t, output, "Major")
	assert.Contains(t, output, "Minor")
	assert.Contains(t, output, "Patch")
	assert.Contains(t, output, "MajorMinorPatch")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

// TestRunVars_DisplaysCustomVariables verifies that custom variables from
// config are included in the vars output.
//
// Why: Users need to see their custom variables alongside built-in ones.
//
// What: Configure custom variables, run "config vars", verify they appear.
func TestRunVars_DisplaysCustomVariables(t *testing.T) {
	// Precondition: temp directory with VERSION file and custom config
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	err = os.WriteFile("VERSION", []byte("1.0.0\n"), 0644)
	require.NoError(t, err)

	// Create config with custom variables
	configContent := `custom:
  AppName: MyTestApp
  Environment: testing
`
	err = os.WriteFile(".versionator.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Action: Execute "config vars"
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"config", "vars"})

	err = rootCmd.Execute()

	// Expected: Command succeeds and custom variables are shown
	assert.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "Custom Variables")
	assert.Contains(t, output, "AppName")
	assert.Contains(t, output, "MyTestApp")

	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}
