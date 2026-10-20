package filter

import (
	"encoding/json"
	"testing"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

func TestParseToolUseResponse(t *testing.T) {
	titles := []string{"Software Engineer", "Product Manager", "Backend Developer"}
	input := titleResults{
		Results: []titleResult{
			{Index: 0, Relevant: true},
			{Index: 1, Relevant: false},
			{Index: 2, Relevant: true},
		},
	}
	inputJSON, _ := json.Marshal(input)

	msg := &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{
			{
				Type:  "tool_use",
				Name:  "classify_titles",
				Input: inputJSON,
			},
		},
	}

	results, err := parseToolUseResponse(msg, titles, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if !results["Software Engineer"] {
		t.Error("expected Software Engineer to be relevant")
	}
	if results["Product Manager"] {
		t.Error("expected Product Manager to be irrelevant")
	}
	if !results["Backend Developer"] {
		t.Error("expected Backend Developer to be relevant")
	}
}

func TestParseToolUseResponseWithOffset(t *testing.T) {
	titles := []string{"DevOps Engineer", "Data Scientist"}
	input := titleResults{
		Results: []titleResult{
			{Index: 100, Relevant: true},
			{Index: 101, Relevant: false},
		},
	}
	inputJSON, _ := json.Marshal(input)

	msg := &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{
			{
				Type:  "tool_use",
				Name:  "classify_titles",
				Input: inputJSON,
			},
		},
	}

	results, err := parseToolUseResponse(msg, titles, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !results["DevOps Engineer"] {
		t.Error("expected DevOps Engineer to be relevant")
	}
	if results["Data Scientist"] {
		t.Error("expected Data Scientist to be irrelevant")
	}
}

func TestParseToolUseResponseNoToolUse(t *testing.T) {
	msg := &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{
			{Type: "text", Text: "some text"},
		},
	}

	_, err := parseToolUseResponse(msg, []string{"test"}, 0)
	if err == nil {
		t.Fatal("expected error for missing tool_use block")
	}
}
