package normalizer

import (
	"reflect"
	"testing"
)

func TestCleanText(t *testing.T) {
	got := CleanText(` <color=Dark>Hello</color>\n  world  `)
	if got != "Hello\nworld" {
		t.Fatalf("CleanText() = %q", got)
	}
}

func TestApplyParamsPreservesMissingValues(t *testing.T) {
	got, warnings := ApplyParams("Deals {0} for {1}s.", []any{[]any{"40.00%"}})
	if got != "Deals 40.00% for {1}s." {
		t.Fatalf("ApplyParams() = %q", got)
	}
	if !reflect.DeepEqual(warnings, []string{"missing parameter {1}"}) {
		t.Fatalf("warnings = %#v", warnings)
	}
}

func TestApplyParamsPlatformAndGenderTags(t *testing.T) {
	got, _ := ApplyParams("{Cus:Ipt,Touch=Tap PC=Press} — {Male=he;Female=she}", nil)
	if got != "Tap — he" {
		t.Fatalf("ApplyParams() = %q", got)
	}
}
