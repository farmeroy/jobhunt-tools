package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"jobhunt/tools/job-monitor/store"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

const chunkSize = 100

type titleResult struct {
	ID       int64 `json:"id"`
	Relevant bool  `json:"relevant"`
}

type titleResults struct {
	Results []titleResult `json:"results"`
}

var classifyTool = anthropic.ToolParam{
	Name:        "classify_titles",
	Description: param.Opt[string]{Value: "Classify job titles as relevant or irrelevant to the candidate's resume"},
	InputSchema: anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"results": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":       map[string]any{"type": "integer", "description": "The posting ID"},
						"relevant": map[string]any{"type": "boolean", "description": "true if the title is relevant to the resume"},
					},
					"required": []string{"id", "relevant"},
				},
			},
		},
		Required: []string{"results"},
	},
}

func FilterTitles(ctx context.Context, client *anthropic.Client, resumeText string, postings []store.Posting) (map[int64]bool, error) {
	if len(postings) == 0 {
		return nil, nil
	}

	results := make(map[int64]bool, len(postings))

	for i := 0; i < len(postings); i += chunkSize {
		end := i + chunkSize
		if end > len(postings) {
			end = len(postings)
		}
		chunk := postings[i:end]

		chunkResults, err := classifyChunk(ctx, client, resumeText, chunk)
		if err != nil {
			return nil, fmt.Errorf("classify chunk %d-%d: %w", i, end, err)
		}
		for id, relevant := range chunkResults {
			results[id] = relevant
		}

		missing := findMissing(chunk, chunkResults)
		if len(missing) > 0 {
			log.Printf("  retry: model skipped %d postings, retrying", len(missing))
			retryResults, err := classifyChunk(ctx, client, resumeText, missing)
			if err != nil {
				log.Printf("  warning: retry failed: %v", err)
			} else {
				for id, relevant := range retryResults {
					results[id] = relevant
				}
				stillMissing := findMissing(missing, retryResults)
				if len(stillMissing) > 0 {
					log.Printf("  warning: %d postings still unclassified after retry", len(stillMissing))
				}
			}
		}
	}

	return results, nil
}

func findMissing(postings []store.Posting, results map[int64]bool) []store.Posting {
	var missing []store.Posting
	for _, p := range postings {
		if _, ok := results[p.ID]; !ok {
			missing = append(missing, p)
		}
	}
	return missing
}

func classifyChunk(ctx context.Context, client *anthropic.Client, resumeText string, postings []store.Posting) (map[int64]bool, error) {
	var listing strings.Builder
	for _, p := range postings {
		fmt.Fprintf(&listing, "- ID %d: %s (at %s)\n", p.ID, p.Title, p.CompanyName)
	}

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: "You are a job relevance classifier. Given a candidate's resume and a list of job titles, classify each title as relevant or irrelevant to the candidate's background and skills.\n\nA title is relevant if the candidate would be a reasonable fit based on their experience. Consider the candidate's technical skills, seniority level, and domain expertise.\n\nClassify every title in the list. Do not skip any."},
		},
		Messages: []anthropic.MessageParam{
			{
				Role: "user",
				Content: []anthropic.ContentBlockParamUnion{
					{OfText: &anthropic.TextBlockParam{Text: fmt.Sprintf("Here is the candidate's resume:\n\n%s\n\nClassify each of these job postings as relevant or irrelevant:\n\n%s", resumeText, listing.String())}},
				},
			},
		},
		Tools: []anthropic.ToolUnionParam{
			{OfTool: &classifyTool},
		},
		ToolChoice: anthropic.ToolChoiceUnionParam{
			OfTool: &anthropic.ToolChoiceToolParam{
				Name: "classify_titles",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic api: %w", err)
	}

	return parseToolUseResponse(resp)
}

func parseToolUseResponse(resp *anthropic.Message) (map[int64]bool, error) {
	for _, block := range resp.Content {
		if block.Type != "tool_use" || block.Name != "classify_titles" {
			continue
		}

		var tr titleResults
		if err := json.Unmarshal(block.Input, &tr); err != nil {
			return nil, fmt.Errorf("parse tool response: %w", err)
		}

		results := make(map[int64]bool, len(tr.Results))
		for _, r := range tr.Results {
			results[r.ID] = r.Relevant
		}
		return results, nil
	}

	return nil, fmt.Errorf("no classify_titles tool use in response")
}
