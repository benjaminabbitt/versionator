---
title: Grammar-Based Parser
description: How versionator uses a formal grammar for version string parsing
sidebar_position: 3
---

# Grammar-Based Parser

Versionator uses a formal EBNF grammar to parse version strings. This page explains the approach, its benefits, and trade-offs.

## Why a Formal Grammar?

Most version parsers use ad-hoc string manipulation or regular expressions. Versionator takes a different approach: we define a formal grammar based on the [SemVer 2.0.0 specification](https://semver.org/) and generate a parser from it.

### The Problem with Ad-Hoc Parsing

Traditional regex-based version parsing has issues:

```go
// Common regex approach - fragile and hard to maintain
var semverRegex = regexp.MustCompile(
    `^v?(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z-.]+)?(\+[0-9A-Za-z-.]+)?$`)
```

Problems:
- **Hard to extend** - Adding support for new formats requires regex surgery
- **Difficult to validate** - Edge cases (leading zeros, empty identifiers) need separate checks
- **Poor error messages** - Regex failures give no context about what went wrong
- **Specification drift** - Easy to deviate from the SemVer spec without noticing

### The Grammar Approach

Instead, we define the grammar formally:

```ebnf
version = [ prefix ], version-core ;
prefix = "v" | "V" ;

version-core = semver | partial-version ;

semver = major, ".", minor, ".", patch, [ pre-release ], [ build-metadata ] ;

pre-release = "-", pre-release-identifier, { ".", pre-release-identifier } ;
pre-release-identifier = alphanumeric-identifier | numeric-identifier ;

numeric-identifier = "0" | positive-digit, { digit } ;
```

This grammar is:
- **Readable** - Maps directly to the SemVer specification
- **Verifiable** - Can be reviewed against the spec
- **Self-documenting** - The grammar IS the documentation

## Implementation

Versionator uses [Participle](https://github.com/alecthomas/participle), a parser generator for Go that builds parsers from struct tags.

### Lexer

The lexer tokenizes input into meaningful units:

```go
var VersionLexer = lexer.MustSimple([]lexer.SimpleRule{
    {Name: "Prefix", Pattern: `[vV]`},
    {Name: "Number", Pattern: `[0-9]+`},
    {Name: "Dot", Pattern: `\.`},
    {Name: "Dash", Pattern: `-`},
    {Name: "Plus", Pattern: `\+`},
    {Name: "Ident", Pattern: `[a-zA-Z][a-zA-Z0-9-]*`},
    {Name: "Mixed", Pattern: `[0-9]+[a-zA-Z][a-zA-Z0-9-]*`},
})
```

### AST

The grammar maps to Go structs:

```go
type Version struct {
    Prefix        string         `parser:"@Prefix?"`
    Core          *VersionCore   `parser:"@@"`
    PreRelease    []*Identifier  `parser:"('-' @@ ('.' @@)* )?"`
    BuildMetadata []*Identifier  `parser:"('+' @@ ('.' @@)* )?"`
    Raw           string
}

type VersionCore struct {
    Major    *int `parser:"@Number"`
    Minor    *int `parser:"('.' @Number)?"`
    Patch    *int `parser:"('.' @Number)?"`
    Revision *int `parser:"('.' @Number)?"`
}
```

### Validation

After parsing, semantic validation ensures SemVer compliance:

- No leading zeros in numeric identifiers (except `0` itself)
- Pre-release identifiers contain only `[0-9A-Za-z-]`
- Build metadata identifiers contain only `[0-9A-Za-z-]`
- Only `v` or `V` prefixes allowed

## Benefits

### 1. Correctness by Construction

The grammar enforces structure that regex can't:

```
1.2.3-alpha.1      ✓ Valid
1.2.3-alpha..1     ✗ Empty identifier (caught by grammar)
1.2.3-alpha.01     ✗ Leading zero (caught by validation)
```

### 2. Better Error Messages

Grammar-based parsing knows exactly where parsing failed:

```
parse error: 1:6: unexpected token "."
              ^
         1.2..3
```

### 3. Easy Extension

Adding support for new formats requires adding grammar rules, not debugging regex:

```ebnf
(* Add 4-component .NET assembly versions *)
assembly-version = major, ".", minor, ".", build, ".", revision ;
```

### 4. Testability

Grammar rules can be tested in isolation:

```go
func TestParse_PreRelease(t *testing.T) {
    tests := []struct{
        input    string
        expected []string
    }{
        {"1.0.0-alpha", []string{"alpha"}},
        {"1.0.0-alpha.1", []string{"alpha", "1"}},
        {"1.0.0-alpha.beta.1", []string{"alpha", "beta", "1"}},
    }
    // ...
}
```

### 5. Documentation as Code

The grammar file serves as executable documentation:

- `docs/grammar/version.ebnf` - Complete formal grammar
- `docs/grammar/railroad.html` - Visual railroad diagrams

## Trade-offs

### Complexity

A grammar-based parser is more complex than a simple regex:

| Approach | Lines of Code | Dependencies |
|----------|---------------|--------------|
| Regex | ~50 | None |
| Grammar | ~300 | participle |

### Build-Time Cost

The parser is constructed at init time, adding ~5ms to startup. This is negligible for CLI usage but measurable.

### Learning Curve

Contributors need to understand:
- EBNF grammar notation
- Participle's struct tag syntax
- Lexer token definitions

## When This Matters

The grammar approach pays off when:

1. **Parsing complex formats** - SemVer has many edge cases
2. **Strict compliance needed** - We claim SemVer 2.0.0 compliance
3. **Multiple formats supported** - Go pseudo-versions, assembly versions, partial versions
4. **Good error messages matter** - CLI users need actionable feedback

## Grammar Reference

The complete grammar is in [`docs/grammar/version.ebnf`](https://github.com/benjaminabbitt/versionator/blob/master/docs/grammar/version.ebnf).

Key rules:

| Rule | Description |
|------|-------------|
| `version` | Top-level: optional prefix + version-core |
| `prefix` | Only `v` or `V` per SemVer convention |
| `semver` | Full Major.Minor.Patch with optional pre-release/metadata |
| `partial-version` | Major only or Major.Minor (defaults missing to 0) |
| `pre-release` | Dash-prefixed, dot-separated identifiers |
| `build-metadata` | Plus-prefixed, dot-separated identifiers |
| `numeric-identifier` | No leading zeros (except "0") |
| `alphanumeric-identifier` | Contains at least one letter |

## Community Use

The EBNF grammar in [`docs/grammar/version.ebnf`](https://github.com/benjaminabbitt/versionator/blob/master/docs/grammar/version.ebnf) is available for the broader SemVer community.

### As Implemented by the Community

This grammar represents SemVer **as implemented by the community**, not just the formal specification. We've tried to capture:

- **SemVer 2.0.0 core** - The official specification
- **Common extensions** - Go module pseudo-versions, npm conventions
- **Practical variations** - Partial versions (`1.2`), optional prefix
- **Adjacent ecosystems** - Microsoft Assembly versions (4-component)

We explicitly chose *not* to make unilateral modifications to the SemVer core. Where the community has established patterns (like the `v` prefix), we document and support them. Where ecosystems have their own conventions (like .NET's 4-component versions), we include them as separate grammar rules rather than mixing them with SemVer.

### Why This Matters

The SemVer specification uses prose descriptions that leave room for interpretation. A formal grammar removes ambiguity and provides a reference that can be:

- **Ported to other languages** - The EBNF notation is parser-generator agnostic
- **Used as a test oracle** - Validate your parser against the grammar
- **Extended for your needs** - Add custom rules while maintaining SemVer compliance
- **Referenced in discussions** - Point to specific grammar rules when debating edge cases

### Contributing

If you find the grammar useful, consider adopting it in your project. If you find bugs, ambiguities, or community conventions we've missed, we welcome issues and pull requests.

### Licensing

The [SemVer 2.0.0 specification](https://semver.org/) by Tom Preston-Werner is licensed under [Creative Commons Attribution 3.0 (CC BY 3.0)](https://creativecommons.org/licenses/by/3.0/).

Our EBNF grammar is a formalization of that specification. The grammar file includes:
- **SemVer rules** - Derived from the CC BY 3.0 specification (attribution preserved)
- **Extensions** - Go pseudo-versions, Assembly versions, partial versions (BSD 3-Clause)
- **Original formalization work** - The EBNF encoding itself (BSD 3-Clause)

**Important:** Our release of this grammar under BSD 3-Clause does *not* remove or supersede the CC BY 3.0 attribution requirements of the underlying SemVer specification. When using the grammar, you must retain attribution to Tom Preston-Werner and the SemVer project as required by CC BY 3.0.

## See Also

- [Version Grammar Explained](./version-grammar) - Plain English guide to version string syntax
- [SemVer 2.0.0 Specification](https://semver.org/spec/v2.0.0.html)
- [VERSION File Format](./version-file)
- [Semantic Versioning Concepts](./semver)
