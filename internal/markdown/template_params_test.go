package markdown

import (
	"reflect"
	"testing"
)

func TestExtractTemplateParameters(t *testing.T) {
	body := `Translate to {{target_language}}:

{{text}}

` + "```" + `
ignore {{secret}}
` + "```"
	got := ExtractTemplateParameters(body)
	want := []string{"target_language", "text"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestExtractTemplateParameters_Dedupes(t *testing.T) {
	got := ExtractTemplateParameters("{{lang}} and again {{lang}}")
	if len(got) != 1 || got[0] != "lang" {
		t.Fatalf("got %v", got)
	}
}
