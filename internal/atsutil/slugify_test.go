package atsutil

import (
	"testing"
)

func TestSlugifySingleWord(t *testing.T) {
	slugs := Slugify("Airbnb")
	if len(slugs) == 0 {
		t.Fatal("expected at least one slug")
	}
	if slugs[0] != "airbnb" {
		t.Errorf("expected first slug to be 'airbnb', got %q", slugs[0])
	}
}

func TestSlugifyMultiWord(t *testing.T) {
	slugs := Slugify("Clipboard Health")
	want := map[string]bool{
		"clipboardhealth": true,
		"clipboard-health": true,
		"clipboard":        true,
	}
	got := map[string]bool{}
	for _, s := range slugs {
		got[s] = true
	}
	for w := range want {
		if !got[w] {
			t.Errorf("expected slug %q in result, got %v", w, slugs)
		}
	}
}

func TestSlugifyWithNumbers(t *testing.T) {
	slugs := Slugify("10x Genomics")
	got := map[string]bool{}
	for _, s := range slugs {
		got[s] = true
	}
	if !got["10xgenomics"] {
		t.Errorf("expected '10xgenomics' in %v", slugs)
	}
	if !got["10x-genomics"] {
		t.Errorf("expected '10x-genomics' in %v", slugs)
	}
}

func TestSlugifyPunctuation(t *testing.T) {
	slugs := Slugify("Company, Inc.")
	got := map[string]bool{}
	for _, s := range slugs {
		got[s] = true
	}
	if !got["company"] {
		t.Errorf("expected 'company' (suffix-stripped) in %v", slugs)
	}
	if !got["companyinc"] {
		t.Errorf("expected 'companyinc' in %v", slugs)
	}
}

func TestSlugifySuffixStripping(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Sapling AI", "sapling"},
		{"Example IO", "example"},
		{"Startup Co", "startup"},
		{"Big Inc", "big"},
		{"Clipboard Health", "clipboard"},
	}
	for _, tc := range tests {
		slugs := Slugify(tc.input)
		got := map[string]bool{}
		for _, s := range slugs {
			got[s] = true
		}
		if !got[tc.expected] {
			t.Errorf("Slugify(%q): expected %q in %v", tc.input, tc.expected, slugs)
		}
	}
}

func TestSlugifyNoDuplicates(t *testing.T) {
	slugs := Slugify("Test")
	seen := map[string]bool{}
	for _, s := range slugs {
		if seen[s] {
			t.Errorf("duplicate slug %q in %v", s, slugs)
		}
		seen[s] = true
	}
}

func TestSlugifyEmptyFields(t *testing.T) {
	slugs := Slugify("  Spaces   Everywhere  ")
	if len(slugs) == 0 {
		t.Fatal("expected at least one slug")
	}
	got := map[string]bool{}
	for _, s := range slugs {
		got[s] = true
	}
	if !got["spaceseverywhere"] {
		t.Errorf("expected 'spaceseverywhere' in %v", slugs)
	}
}
