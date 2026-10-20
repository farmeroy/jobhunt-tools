package filter

import (
	"encoding/json"
	"testing"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

func TestParseToolUseResponse(t *testing.T) {
	input := titleResults{
		Results: []titleResult{
			{ID: 1, Relevant: true},
			{ID: 2, Relevant: false},
			{ID: 3, Relevant: true},
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

	results, err := parseToolUseResponse(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if !results[1] {
		t.Error("expected ID 1 to be relevant")
	}
	if results[2] {
		t.Error("expected ID 2 to be irrelevant")
	}
	if !results[3] {
		t.Error("expected ID 3 to be relevant")
	}
}

func TestParseToolUseResponseNoToolUse(t *testing.T) {
	msg := &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{
			{Type: "text", Text: "some text"},
		},
	}

	_, err := parseToolUseResponse(msg)
	if err == nil {
		t.Fatal("expected error for missing tool_use block")
	}
}

func TestChunking(t *testing.T) {
	if chunkSize != 100 {
		t.Fatalf("expected chunk size 100, got %d", chunkSize)
	}
}
