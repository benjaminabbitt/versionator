// Package parser provides a grammar-based version string parser using participle.
// The grammar supports SemVer 2.0.0, Go module versions, Microsoft assembly versions,
// and various extensions as documented in docs/grammar/version.ebnf.
package parser

import (
	"github.com/alecthomas/participle/v2/lexer"
)

// VersionLexer defines the lexical tokens for version strings.
// Token order matters - more specific patterns must come before general ones.
//
// The lexer uses several token types:
//   - Number: pure digits (for version components)
//   - Prefix: v or V (for version prefix)
//   - Ident: alphanumeric with optional dashes (for pre-release/metadata)
//   - Mixed: digit-starting alphanumeric (like "5114f85")
//   - Dashes: two or more dashes (valid SemVer identifier like "--")
//
// Per SemVer 2.0.0, identifiers can be purely dashes (e.g., "1.0.0-x-y-z.--")
var VersionLexer = lexer.MustSimple([]lexer.SimpleRule{
	// Whitespace (ignored)
	{Name: "Whitespace", Pattern: `[ \t]+`},

	// Separators
	{Name: "Dot", Pattern: `\.`},
	{Name: "Plus", Pattern: `\+`},

	// Dashes: two or more dashes (valid identifier per SemVer 2.0.0)
	// Must come BEFORE single Dash to match "--" etc as identifiers
	{Name: "Dashes", Pattern: `--+`},

	// Single dash: pre-release separator or part of identifier
	{Name: "Dash", Pattern: `-`},

	// Prefix: single v or V (case-insensitive)
	{Name: "Prefix", Pattern: `[vV]`},

	// Mixed alphanumeric: starts with digit but contains letters (like "5114f85", "21AF26D3")
	// Must come before Number to match these first
	{Name: "Mixed", Pattern: `[0-9]+[a-zA-Z][a-zA-Z0-9-]*`},

	// Pure alphanumeric: starts with letter (like "alpha", "rc1", "beta-2")
	{Name: "Ident", Pattern: `[a-zA-Z][a-zA-Z0-9-]*`},

	// Pure numeric: digits only (for version components like "1", "23", "456")
	{Name: "Number", Pattern: `[0-9]+`},
})
