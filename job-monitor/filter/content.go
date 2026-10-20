package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

type ContentPosting struct {
	ID          int64
	Title       string
	Description string
}

type ContentResult struct {
	PostingID int64
	Score     float64
	Reasoning string
}

type scoreOutput struct {
	Score     int    `json:"score"`
	Reasoning string `json:"reasoning"`
}

var scoreTool = anthropic.ToolParam{
	Name:        "score_posting",
	Description: param.Opt[string]{Value: "Score how well the candidate matches this role"},
	InputSchema: anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"score": map[string]any{
				"type":        "integer",
				"description": "1-10 confidence that you would pass this candidate to the hiring manager. 1 = no fit, 10 = perfect fit.",
				"minimum":     1,
				"maximum":     10,
			},
			"reasoning": map[string]any{
				"type":        "string",
				"description": "Brief reasoning: key requirements met and unmet.",
			},
		},
		Required: []string{"score", "reasoning"},
	},
}

func ScoreContent(ctx context.Context, client *anthropic.Client, resumeText string, postings []ContentPosting) ([]ContentResult, *Usage, error) {
	usage := &Usage{}

	var skipped int
	var results []ContentResult

	for _, p := range postings {
		if p.Description == "" {
			skipped++
			continue
		}

		result, resp, err := scorePosting(ctx, client, resumeText, p)
		if err != nil {
			log.Printf("  error scoring posting %d (%s): %v", p.ID, p.Title, err)
			continue
		}
		usage.add(resp)
		results = append(results, *result)
		log.Printf("  [%.0f/10] %s", result.Score, p.Title)
	}

	if skipped > 0 {
		log.Printf("  skipped %d postings with empty descriptions", skipped)
	}

	return results, usage, nil
}

func scorePosting(ctx context.Context, client *anthropic.Client, resumeText string, p ContentPosting) (*ContentResult, *anthropic.Message, error) {
	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: fmt.Sprintf("You are a recruiter screening candidates for the following role:\n\nTitle: %s\n\nJob Description:\n%s\n\nYou are thorough and slightly skeptical. You look for concrete evidence of fit, not just keyword overlap.", p.Title, p.Description)},
		},
		Messages: []anthropic.MessageParam{
			{
				Role: "user",
				Content: []anthropic.ContentBlockParamUnion{
					{OfText: &anthropic.TextBlockParam{Text: fmt.Sprintf("Here is a candidate's resume. Would you pass this candidate on to the hiring manager?\n\nResume:\n\n%s", resumeText)}},
				},
			},
		},
		Tools: []anthropic.ToolUnionParam{
			{OfTool: &scoreTool},
		},
		ToolChoice: anthropic.ToolChoiceUnionParam{
			OfTool: &anthropic.ToolChoiceToolParam{
				Name: "score_posting",
			},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic api: %w", err)
	}

	for _, block := range resp.Content {
		if block.Type != "tool_use" || block.Name != "score_posting" {
			continue
		}

		var out scoreOutput
		if err := json.Unmarshal(block.Input, &out); err != nil {
			return nil, nil, fmt.Errorf("parse tool response: %w", err)
		}

		return &ContentResult{
			PostingID: p.ID,
			Score:     float64(out.Score),
			Reasoning: out.Reasoning,
		}, resp, nil
	}

	return nil, nil, fmt.Errorf("no score_posting tool use in response")
}
