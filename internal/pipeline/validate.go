package pipeline

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/kiwifs/kiwifs/internal/config"
	"github.com/kiwifs/kiwifs/internal/markdown"
	"github.com/kiwifs/kiwifs/internal/storage"
)

// WriteRuleValidator applies [[validate_write]] rules from config.toml.
type WriteRuleValidator struct {
	store storage.Storage
	rules []config.ValidateWriteRuleConfig
}

// NewWriteRuleValidator builds a validator for the configured rules.
func NewWriteRuleValidator(store storage.Storage, rules []config.ValidateWriteRuleConfig) *WriteRuleValidator {
	return &WriteRuleValidator{store: store, rules: rules}
}

// Validate checks configured rules against the existing file (when present).
func (v *WriteRuleValidator) Validate(ctx context.Context, path string, newContent []byte, kind WriteKind) error {
	if v == nil || len(v.rules) == 0 {
		return nil
	}
	existing, err := v.store.Read(ctx, path)
	if err != nil || len(existing) == 0 {
		return nil
	}
	fm, err := markdown.Frontmatter(existing)
	if err != nil || len(fm) == 0 {
		return nil
	}
	for _, rule := range v.rules {
		if !matchFrontmatter(fm, rule.Match) {
			continue
		}
		switch rule.Reject {
		case "overwrite":
			if kind == WriteKindPut {
				return rejectedWrite(rule.Message)
			}
		case "body_change":
			if bodyChanged(existing, newContent) {
				return rejectedWrite(rule.Message)
			}
		default:
			return fmt.Errorf("validate_write rule %q: unknown reject %q", rule.Name, rule.Reject)
		}
	}
	return nil
}

func rejectedWrite(message string) error {
	if message == "" {
		return ErrWriteRejected
	}
	return fmt.Errorf("%w: %s", ErrWriteRejected, message)
}

func wrapValidateWriteErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrWriteRejected) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrValidationFailed, err)
}

func matchFrontmatter(fm map[string]any, match config.ValidateWriteMatchConfig) bool {
	if match.Frontmatter == "" {
		return false
	}
	val, ok := fm[match.Frontmatter]
	if !ok {
		return false
	}
	if match.Value != "" {
		return frontmatterValueEquals(val, match.Value)
	}
	for _, wanted := range match.Values {
		if frontmatterValueEquals(val, wanted) {
			return true
		}
	}
	return false
}

func frontmatterValueEquals(actual any, expected string) bool {
	switch v := actual.(type) {
	case string:
		return v == expected
	case bool:
		switch expected {
		case "true":
			return v
		case "false":
			return !v
		default:
			return strconv.FormatBool(v) == expected
		}
	case float64:
		return fmt.Sprint(v) == expected || strconv.FormatFloat(v, 'f', -1, 64) == expected
	case int:
		return strconv.Itoa(v) == expected
	default:
		return fmt.Sprint(v) == expected
	}
}

func bodyChanged(existing, updated []byte) bool {
	_, oldBody, oldErr := markdown.SplitFrontmatter(existing)
	_, newBody, newErr := markdown.SplitFrontmatter(updated)
	if oldErr != nil || newErr != nil {
		return !bytes.Equal(existing, updated)
	}
	return !bytes.Equal(oldBody, newBody)
}
