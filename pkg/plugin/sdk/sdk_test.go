package sdk

import "testing"

func TestGetVar_KeyExists_ReturnsValue(t *testing.T) {
	vars := map[string]string{"Version": "1.2.3"}

	result := GetVar(vars, "Version", "default")

	if result != "1.2.3" {
		t.Errorf("expected '1.2.3', got '%s'", result)
	}
}

func TestGetVar_KeyMissing_ReturnsDefault(t *testing.T) {
	vars := map[string]string{"Other": "value"}

	result := GetVar(vars, "Version", "0.0.0")

	if result != "0.0.0" {
		t.Errorf("expected '0.0.0', got '%s'", result)
	}
}

func TestGetVar_EmptyValue_ReturnsDefault(t *testing.T) {
	vars := map[string]string{"Version": ""}

	result := GetVar(vars, "Version", "default")

	if result != "default" {
		t.Errorf("expected 'default' for empty value, got '%s'", result)
	}
}

func TestGetVar_NilMap_ReturnsDefault(t *testing.T) {
	result := GetVar(nil, "Version", "fallback")

	if result != "fallback" {
		t.Errorf("expected 'fallback' for nil map, got '%s'", result)
	}
}

func TestGetVar_EmptyDefault_ReturnsEmptyString(t *testing.T) {
	vars := map[string]string{}

	result := GetVar(vars, "Missing", "")

	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestGetNumericVar_KeyExists_ReturnsValue(t *testing.T) {
	vars := map[string]string{"Major": "2"}

	result := GetNumericVar(vars, "Major")

	if result != "2" {
		t.Errorf("expected '2', got '%s'", result)
	}
}

func TestGetNumericVar_KeyMissing_ReturnsZero(t *testing.T) {
	vars := map[string]string{}

	result := GetNumericVar(vars, "Major")

	if result != "0" {
		t.Errorf("expected '0', got '%s'", result)
	}
}

func TestGetNumericVar_EmptyValue_ReturnsZero(t *testing.T) {
	vars := map[string]string{"Minor": ""}

	result := GetNumericVar(vars, "Minor")

	if result != "0" {
		t.Errorf("expected '0' for empty value, got '%s'", result)
	}
}

func TestGetNumericVar_NilMap_ReturnsZero(t *testing.T) {
	result := GetNumericVar(nil, "Patch")

	if result != "0" {
		t.Errorf("expected '0' for nil map, got '%s'", result)
	}
}

func TestGetNumericVar_NonNumericValue_ReturnsAsIs(t *testing.T) {
	// GetNumericVar doesn't validate - it's the caller's responsibility
	vars := map[string]string{"Major": "abc"}

	result := GetNumericVar(vars, "Major")

	if result != "abc" {
		t.Errorf("expected 'abc' (no validation), got '%s'", result)
	}
}
