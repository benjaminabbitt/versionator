package update

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

// bareKeyPath matches a selector made only of dotted bare TOML keys, such as
// "workspace.package.version". Such a path names one value of the document,
// which setTOMLString finds by the document's own structure.
var bareKeyPath = regexp.MustCompile(`^[A-Za-z0-9_-]+(\.[A-Za-z0-9_-]+)*$`)

// setTOMLString returns raw with the string at path replaced by newValue,
// every other byte as it was (comments, ordering, spacing, quote style).
//
// A path of dotted bare keys is located with go-toml's parser, so the value is
// the one at that path and never a lookalike in another table (a dependency
// pinned to the same version). Any other selector syntax keeps the earlier
// behaviour: the value is read through dasel and its first quoted occurrence
// in the file is replaced.
func (p *DaselFileParser) setTOMLString(raw []byte, path, newValue string) ([]byte, error) {
	if !bareKeyPath.MatchString(path) {
		return p.replaceFirstTOMLOccurrence(raw, path, newValue)
	}

	start, end, err := locateTOMLString(raw, strings.Split(path, "."))
	if err != nil {
		return nil, err
	}

	old := raw[start:end]
	quoted := quoteTOMLLike(old, newValue)
	if bytes.Equal(old, quoted) {
		return raw, nil
	}

	out := append([]byte{}, raw[:start]...)
	out = append(out, quoted...)
	return append(out, raw[end:]...), nil
}

// locateTOMLString returns the byte range, quotes included, of the single-line
// string value at keys. It is an error for the path to name nothing, or to
// name something other than a single-line string (a table, a number, or an
// inherited `version.workspace = true`).
func locateTOMLString(raw []byte, keys []string) (int, int, error) {
	path := strings.Join(keys, ".")
	parser := unstable.Parser{}
	parser.Reset(raw)

	var table []string
	for parser.NextExpression() {
		expr := parser.Expression()
		switch expr.Kind {
		case unstable.Table:
			table = keyParts(expr.Key())
			if isPrefix(keys, table) {
				return 0, 0, fmt.Errorf("%s: %s is a table", ErrValueNotString, path)
			}
		case unstable.ArrayTable:
			// [[array]] entries are not addressable by a dotted path: the "[]"
			// segment is no bare key, so no path reaches a key below it.
			table = append(keyParts(expr.Key()), "[]")
		case unstable.KeyValue:
			start, end, found, err := matchKeyValue(expr, table, keys, path)
			if err != nil {
				return 0, 0, err
			}
			if found {
				// A multi-line string ("""...""" or '''...''') is not rewritten.
				if q := raw[start:end]; bytes.HasPrefix(q, []byte(`"""`)) || bytes.HasPrefix(q, []byte("'''")) {
					return 0, 0, fmt.Errorf("%s: %s is a multi-line string", ErrValueNotString, path)
				}
				return start, end, nil
			}
		}
	}
	if err := parser.Error(); err != nil {
		return 0, 0, fmt.Errorf("%s: %w", ErrFileParseFailed, err)
	}
	return 0, 0, fmt.Errorf("%s: %s", ErrPathNotFound, path)
}

// matchKeyValue checks one key = value expression (inside table) against keys,
// descending into an inline table that lies on the path.
func matchKeyValue(expr *unstable.Node, table, keys []string, path string) (int, int, bool, error) {
	full := append(append([]string{}, table...), keyParts(expr.Key())...)
	value := expr.Value()

	if equalKeys(full, keys) {
		if value.Kind != unstable.String {
			return 0, 0, false, fmt.Errorf("%s: %s is a %s", ErrValueNotString, path, value.Kind)
		}
		start := int(value.Raw.Offset)
		return start, start + int(value.Raw.Length), true, nil
	}
	if isPrefix(keys, full) {
		// keys names a table here, made by dotted keys (version.workspace = true).
		return 0, 0, false, fmt.Errorf("%s: %s is a table", ErrValueNotString, path)
	}
	if value.Kind == unstable.InlineTable {
		children := value.Children()
		// An inline table's children are all key = value expressions; one off
		// the path matches nothing, so every inline table may be searched.
		for children.Next() {
			start, end, found, err := matchKeyValue(children.Node(), full, keys, path)
			if err != nil || found {
				return start, end, found, err
			}
		}
	}
	return 0, 0, false, nil
}

// quoteTOMLLike renders value as a TOML string in the quote style of old: a
// literal '...' string stays literal when value can be one, and anything else
// is a basic "..." string, escaped.
func quoteTOMLLike(old []byte, value string) []byte {
	// old is a located string, so it holds at least its two quotes.
	if old[0] == '\'' && !strings.ContainsAny(value, "'\n\r") {
		return []byte("'" + value + "'")
	}
	return []byte(`"` + basicStringEscaper.Replace(value) + `"`)
}

var basicStringEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

func keyParts(it unstable.Iterator) []string {
	var parts []string
	for it.Next() {
		parts = append(parts, string(it.Node().Data))
	}
	return parts
}

func equalKeys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// isPrefix reports whether keys begins with prefix (or equals it).
func isPrefix(prefix, keys []string) bool {
	return len(prefix) <= len(keys) && equalKeys(prefix, keys[:len(prefix)])
}

// replaceFirstTOMLOccurrence is the earlier text replacement, kept for
// selectors that are not plain dotted keys: the value at path is read through
// dasel, and its first quoted occurrence in the file is replaced.
func (p *DaselFileParser) replaceFirstTOMLOccurrence(raw []byte, path, newValue string) ([]byte, error) {
	var data map[string]any
	if err := toml.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFileParseFailed, err)
	}

	oldValue, err := p.Select(data, path)
	if err != nil {
		return nil, err
	}

	oldStr := fmt.Sprintf("%v", oldValue)
	if oldStr == newValue {
		return raw, nil
	}

	newQuoted := []byte(`"` + newValue + `"`)
	result := bytes.Replace(raw, []byte(`"`+oldStr+`"`), newQuoted, 1)
	if bytes.Equal(result, raw) {
		result = bytes.Replace(raw, []byte(`'`+oldStr+`'`), newQuoted, 1)
	}
	if bytes.Equal(result, raw) {
		return nil, fmt.Errorf("could not find value %q to replace", oldStr)
	}
	return result, nil
}

// cargoVersionPattern is SemVer 2.0.0 (semver.org's own expression), which is
// what Cargo accepts as a package version: exactly Major.Minor.Patch, an
// optional -pre-release and +build metadata, and no "v" prefix.
var cargoVersionPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

func cargoVersionValid(v string) bool {
	return cargoVersionPattern.MatchString(v)
}

// isCargoVersionKey reports whether an update writes a Cargo manifest's
// version: `package.version` or `workspace.package.version` of a file named
// Cargo.toml. Only those are held to Cargo's version rules.
func isCargoVersionKey(file, path string) bool {
	if filepath.Base(file) != "Cargo.toml" {
		return false
	}
	return path == "package.version" || path == "workspace.package.version"
}
