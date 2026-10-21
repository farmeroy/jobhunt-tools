package filter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"golang.org/x/sync/errgroup"
)

type ContentPosting struct {
	ID          int64
	Title       string
	Location    string
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

type ScoreResult struct {
	Scored []ContentResult
	Errors int
}

func ScoreContent(ctx context.Context, client *anthropic.Client, resumeText string, allowedLocations []string, postings []ContentPosting) (*ScoreResult, *Usage, error) {
	usage := &Usage{}

	var toScore []ContentPosting
	var skipped int
	for _, p := range postings {
		if p.Description == "" {
			skipped++
			continue
		}
		toScore = append(toScore, p)
	}
	if skipped > 0 {
		log.Printf("  skipped %d postings with empty descriptions", skipped)
	}

	resultsCh := make(chan ContentResult, len(toScore))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for _, p := range toScore {
		g.Go(func() error {
			result, resp, err := scorePosting(gctx, client, resumeText, allowedLocations, p)
			if err != nil {
				var apierr *anthropic.Error
				if errors.As(err, &apierr) && apierr.StatusCode >= 400 && apierr.StatusCode < 500 && apierr.StatusCode != 429 {
					return fmt.Errorf("posting %d (%s): %w", p.ID, p.Title, err)
				}
				log.Printf("  error scoring posting %d (%s): %v", p.ID, p.Title, err)
				return nil
			}
			usage.add(resp)
			log.Printf("  [%.0f/10] %s", result.Score, p.Title)
			resultsCh <- *result
			return nil
		})
	}

	err := g.Wait()
	close(resultsCh)

	results := make([]ContentResult, 0, len(toScore))
	for r := range resultsCh {
		results = append(results, r)
	}

	if err != nil {
		return nil, usage, err
	}

	return &ScoreResult{
		Scored: results,
		Errors: len(toScore) - len(results),
	}, usage, nil
}

func scorePosting(ctx context.Context, client *anthropic.Client, resumeText string, allowedLocations []string, p ContentPosting) (*ContentResult, *anthropic.Message, error) {
	locationConstraint := ""
	if len(allowedLocations) > 0 {
		locationConstraint = fmt.Sprintf("\n\nLocation constraint: The candidate can only accept roles in these locations: %s. If the posting location or job description indicates the role requires a location outside this list, that is a disqualifying mismatch — score 1. Use geographic reasoning: \"Remote California\" is within \"Remote US\", but \"Remote Canada\" is not.", strings.Join(allowedLocations, ", "))
	}

	postingLocation := ""
	if p.Location != "" {
		postingLocation = fmt.Sprintf("\nPosting Location: %s", p.Location)
	}

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: fmt.Sprintf("You are a recruiter screening candidates for the following role:\n\nTitle: %s%s\n\nJob Description:\n%s\n\nYou are thorough and slightly skeptical. You look for concrete evidence of fit, not just keyword overlap.%s", p.Title, postingLocation, p.Description, locationConstraint)},
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
