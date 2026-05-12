package views

import (
	"testing"
)

func TestEvalFormulaArithmetic(t *testing.T) {
	row := map[string]any{
		"a": 10.0,
		"b": 5.0,
	}

	tests := []struct {
		expr     string
		expected float64
	}{
		{"a + b", 15.0},
		{"a - b", 5.0},
		{"a * b", 50.0},
		{"a / b", 2.0},
	}

	for _, tt := range tests {
		result, err := EvalFormula(tt.expr, row)
		if err != nil {
			t.Errorf("EvalFormula(%q) error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("EvalFormula(%q) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvalFormulaLength(t *testing.T) {
	row := map[string]any{
		"text": "hello",
		"list": []any{1, 2, 3},
	}

	tests := []struct {
		expr     string
		expected int
	}{
		{"length(text)", 5},
		{"length(list)", 3},
	}

	for _, tt := range tests {
		result, err := EvalFormula(tt.expr, row)
		if err != nil {
			t.Errorf("EvalFormula(%q) error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("EvalFormula(%q) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvalFormulaRound(t *testing.T) {
	row := map[string]any{
		"value": 3.14159,
	}

	tests := []struct {
		expr     string
		expected float64
	}{
		{"round(value)", 3.0},
		{"round(value, 2)", 3.14},
	}

	for _, tt := range tests {
		result, err := EvalFormula(tt.expr, row)
		if err != nil {
			t.Errorf("EvalFormula(%q) error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("EvalFormula(%q) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvalFormulaIf(t *testing.T) {
	row := map[string]any{
		"status": "done",
		"empty":  "",
	}

	tests := []struct {
		expr     string
		expected string
	}{
		{`if(status, "yes", "no")`, "yes"},
		{`if(empty, "yes", "no")`, "no"},
	}

	for _, tt := range tests {
		result, err := EvalFormula(tt.expr, row)
		if err != nil {
			t.Errorf("EvalFormula(%q) error: %v", tt.expr, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("EvalFormula(%q) = %v, want %v", tt.expr, result, tt.expected)
		}
	}
}

func TestEvalFormulaFieldReference(t *testing.T) {
	row := map[string]any{
		"count": 42.0,
	}

	result, err := EvalFormula("count", row)
	if err != nil {
		t.Fatalf("EvalFormula error: %v", err)
	}
	if result != 42.0 {
		t.Errorf("expected 42.0, got %v", result)
	}
}

func TestEvalFormulaLiteral(t *testing.T) {
	row := map[string]any{}

	result, err := EvalFormula("100", row)
	if err != nil {
		t.Fatalf("EvalFormula error: %v", err)
	}
	if result != 100.0 {
		t.Errorf("expected 100.0, got %v", result)
	}
}

func TestEvalFormulaDivisionByZero(t *testing.T) {
	row := map[string]any{
		"a": 10.0,
	}

	_, err := EvalFormula("a / 0", row)
	if err == nil {
		t.Error("expected division by zero error")
	}
}

// --- Additional edge case tests ---

func TestFormulaOperatorPrecedence(t *testing.T) {
	row := map[string]any{}
	result, err := EvalFormula("2 + 3 * 4", row)
	if err != nil {
		t.Fatal(err)
	}
	f := result.(float64)
	if f != 14.0 {
		t.Errorf("2 + 3 * 4 = %v, want 14", f)
	}
}

func TestFormulaLeftAssociativity(t *testing.T) {
	row := map[string]any{}
	result, err := EvalFormula("10 - 3 - 2", row)
	if err != nil {
		t.Fatal(err)
	}
	f := result.(float64)
	if f != 5.0 {
		t.Errorf("10 - 3 - 2 = %v, want 5", f)
	}
}

func TestFormulaMixedPrecedence(t *testing.T) {
	row := map[string]any{}
	result, err := EvalFormula("1 + 2 * 3 + 4", row)
	if err != nil {
		t.Fatal(err)
	}
	f := result.(float64)
	if f != 11.0 {
		t.Errorf("1 + 2 * 3 + 4 = %v, want 11", f)
	}
}

func TestFormulaNestedRound(t *testing.T) {
	row := map[string]any{"x": 3.14159}
	result, err := EvalFormula("round(x, 2)", row)
	if err != nil {
		t.Fatal(err)
	}
	if result.(float64) != 3.14 {
		t.Errorf("round(3.14159, 2) = %v, want 3.14", result)
	}
}

func TestFormulaEmptyString(t *testing.T) {
	_, err := EvalFormula("", map[string]any{})
	if err == nil {
		t.Error("expected error for empty expression")
	}
}

func TestFormulaFieldReference(t *testing.T) {
	row := map[string]any{"name": "hello"}
	result, err := EvalFormula("name", row)
	if err != nil {
		t.Fatal(err)
	}
	if result != "hello" {
		t.Errorf("field reference = %v, want 'hello'", result)
	}
}
