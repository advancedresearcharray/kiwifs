//go:build onnx

package embed

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTokenizerSafeMalformedJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		content         string
		wantErrContains string
	}{
		// sugarme/tokenizer panics on structurally incomplete JSON; loadTokenizerSafe recovers.
		{name: "empty object", content: "{}", wantErrContains: "malformed or incompatible"},
		{name: "null", content: "null", wantErrContains: "malformed or incompatible"},
		{name: "missing model field", content: `{"version":"1.0"}`, wantErrContains: "malformed or incompatible"},
		{name: "model null", content: `{"model": null}`, wantErrContains: "malformed or incompatible"},
		{name: "model empty object", content: `{"model": {}}`, wantErrContains: "malformed or incompatible"},
		// syntactically invalid or empty input fails in FromFile before panic recovery.
		{name: "empty file", content: "", wantErrContains: "load tokenizer"},
		{name: "empty array", content: "[]", wantErrContains: "load tokenizer"},
		{name: "truncated JSON", content: `{"model":`, wantErrContains: "load tokenizer"},
		{name: "invalid syntax", content: `{not json}`, wantErrContains: "load tokenizer"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "tokenizer.json")
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}

			_, err := loadTokenizerSafe(path)
			if err == nil {
				t.Fatal("loadTokenizerSafe succeeded on malformed tokenizer.json")
			}
			errMsg := err.Error()
			if !strings.Contains(errMsg, tc.wantErrContains) {
				t.Fatalf("error %q does not contain %q", err, tc.wantErrContains)
			}
			if tc.wantErrContains == "malformed or incompatible" && !strings.Contains(errMsg, path) {
				t.Fatalf("panic-recovery error %q does not include tokenizer path %q", err, path)
			}
		})
	}
}
