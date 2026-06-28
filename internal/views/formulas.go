package views

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// EvalFormula evaluates a simple formula expression against a row of data.
// Supported operations: +, -, *, /, length(), round(), if()
func EvalFormula(expr string, row map[string]any) (any, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Check for function calls
	if strings.Contains(expr, "(") {
		return evalFunction(expr, row)
	}

	// Split on lowest-precedence operator (+ and - before * and /) scanning
	// right-to-left so that left-associativity is preserved (e.g. 3-2-1 = 0).
	// Parenthesised sub-expressions and function calls are skipped.
	if result, err := splitOnOperator(expr, row, []byte{'+', '-'}); err == nil {
		return result, nil
	}
	if result, err := splitOnOperator(expr, row, []byte{'*', '/'}); err == nil {
		return result, nil
	}

	// Check if it's a field reference
	if val, ok := row[expr]; ok {
		return val, nil
	}

	// Try to parse as number
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num, nil
	}

	// Try to parse as string literal (quoted)
	if (strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`)) ||
		(strings.HasPrefix(expr, `'`) && strings.HasSuffix(expr, `'`)) {
		return expr[1 : len(expr)-1], nil
	}

	return nil, fmt.Errorf("unknown expression: %s", expr)
}

func evalFunction(expr string, row map[string]any) (any, error) {
	expr = strings.TrimSpace(expr)

	// Extract function name and arguments
	openIdx := strings.Index(expr, "(")
	closeIdx := strings.LastIndex(expr, ")")
	if openIdx == -1 || closeIdx == -1 || closeIdx < openIdx {
		return nil, fmt.Errorf("invalid function syntax: %s", expr)
	}

	funcName := strings.ToLower(strings.TrimSpace(expr[:openIdx]))
	argsStr := strings.TrimSpace(expr[openIdx+1 : closeIdx])

	switch funcName {
	case "length":
		arg, err := EvalFormula(argsStr, row)
		if err != nil {
			return nil, err
		}
		return length(arg), nil

	case "round":
		args := splitArgs(argsStr)
		if len(args) < 1 {
			return nil, fmt.Errorf("round requires at least 1 argument")
		}
		val, err := EvalFormula(args[0], row)
		if err != nil {
			return nil, err
		}
		precision := 0
		if len(args) > 1 {
			precVal, err := EvalFormula(args[1], row)
			if err != nil {
				return nil, err
			}
			precision = int(toFloat(precVal))
		}
		return round(toFloat(val), precision), nil

	case "if":
		args := splitArgs(argsStr)
		if len(args) != 3 {
			return nil, fmt.Errorf("if requires 3 arguments: condition, true_value, false_value")
		}
		// Simple condition evaluation (field != empty)
		condField := strings.TrimSpace(args[0])
		val, ok := row[condField]
		if ok && val != nil && val != "" {
			return EvalFormula(args[1], row)
		}
		return EvalFormula(args[2], row)

	default:
		return nil, fmt.Errorf("unknown function: %s", funcName)
	}
}

// splitOnOperator scans expr right-to-left for any operator in ops,
// respecting parenthesis depth and quoted strings. Returns the evaluated
// result if an operator was found, otherwise returns a sentinel error so
// the caller can try the next precedence level.
var errNoSplit = fmt.Errorf("no split")

func splitOnOperator(expr string, row map[string]any, ops []byte) (any, error) {
	depth := 0
	inQuotes := false
	runes := []rune(expr)
	// Scan right-to-left for left-associativity.
	for i := len(runes) - 1; i > 0; i-- {
		ch := runes[i]
		if ch == '"' || ch == '\'' {
			inQuotes = !inQuotes
		}
		if inQuotes {
			continue
		}
		if ch == ')' {
			depth++
		}
		if ch == '(' {
			depth--
		}
		if depth != 0 {
			continue
		}
		for _, op := range ops {
			if byte(ch) == op {
				// Avoid matching unary minus at position 0 or after another operator.
				if op == '-' && i > 0 {
					prev := runes[i-1]
					if prev == '+' || prev == '-' || prev == '*' || prev == '/' || prev == '(' {
						continue
					}
				}
				left := strings.TrimSpace(string(runes[:i]))
				right := strings.TrimSpace(string(runes[i+1:]))
				if left == "" || right == "" {
					continue
				}
				leftVal, err := EvalFormula(left, row)
				if err != nil {
					return nil, err
				}
				rightVal, err := EvalFormula(right, row)
				if err != nil {
					return nil, err
				}
				return evalArithmetic(leftVal, rightVal, string(ch))
			}
		}
	}
	return nil, errNoSplit
}

func evalArithmetic(left, right any, op string) (any, error) {
	l := toFloat(left)
	r := toFloat(right)

	switch op {
	case "+":
		return l + r, nil
	case "-":
		return l - r, nil
	case "*":
		return l * r, nil
	case "/":
		if r == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return l / r, nil
	default:
		return nil, fmt.Errorf("unknown operator: %s", op)
	}
}

func toFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

func length(v any) int {
	switch val := v.(type) {
	case string:
		return len(val)
	case []any:
		return len(val)
	default:
		return 0
	}
}

func round(val float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(val*multiplier) / multiplier
}

func splitArgs(argsStr string) []string {
	var args []string
	var current strings.Builder
	depth := 0
	inQuotes := false

	for _, ch := range argsStr {
		if ch == '"' || ch == '\'' {
			inQuotes = !inQuotes
		}
		if !inQuotes && ch == '(' {
			depth++
		}
		if !inQuotes && ch == ')' {
			depth--
		}
		if !inQuotes && depth == 0 && ch == ',' {
			args = append(args, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		current.WriteRune(ch)
	}
	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}
	return args
}
