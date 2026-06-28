package orders

import (
	"regexp"
	"testing"
)

var referenceRegex = regexp.MustCompile(`^RUE-[0-9A-F]{8}$`)

func TestGenerateReference_Format(t *testing.T) {
	ref, err := GenerateReference()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !referenceRegex.MatchString(ref) {
		t.Errorf("reference %q does not match %s", ref, referenceRegex)
	}
}

func TestGenerateReference_Unique(t *testing.T) {
	a, err := GenerateReference()
	if err != nil {
		t.Fatalf("generate a: %v", err)
	}
	b, err := GenerateReference()
	if err != nil {
		t.Fatalf("generate b: %v", err)
	}
	if a == b {
		t.Errorf("expected two calls to differ, both got %q", a)
	}
}
