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
	PostingID    int64
	Score        float64
	Met          []string
	Unmet        []string
	Dealbreakers []string
}

func (r ContentResult) FormatReasoning() string {
	var parts []string
	for _, s := range r.Met {
		parts = append(parts, "+ "+s)
	}
	for _, s := range r.Unmet {
		parts = append(parts, "- "+s)
	}
	for _, s := range r.Dealbreakers {
		parts = append(parts, "! "+s)
	}
	return strings.Join(parts, " | ")
}

type scoreOutput struct {
	Score        int      `json:"score"`
	Met          []string `json:"met"`
	Unmet        []string `json:"unmet"`
	Dealbreakers []string `json:"dealbreakers"`
}

var scoreTool = anthropic.ToolParam{
	Name:        "score_posting",
	Description: param.Opt[string]{Value: "Score how well the candidate matches this role. You MUST call this tool with your assessment."},
	Strict:      param.Opt[bool]{Value: true},
	InputSchema: anthropic.ToolInputSchemaParam{
		ExtraFields: map[string]any{"additionalProperties": false},
		Properties: map[string]any{
			"score": map[string]any{
				"type":        "integer",
				"enum":        []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				"description": "1-10 likelihood the candidate gets an interview. 9-10: meets all key requirements. 7-8: most requirements met, minor gaps. 4-6: mixed fit. 1-3: has a dealbreaker.",
			},
			"met": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Key requirements the candidate meets. Each item: one requirement, brief evidence.",
			},
			"unmet": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Key requirements the candidate does not meet. Each item: one requirement.",
			},
			"dealbreakers": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Hard blockers that should force score to 3 or below. ONLY these qualify: location mismatch, total experience gap >2 years below stated minimum, or missing a core technical skill that defines the role (e.g. embedded C role but candidate has only web experience). Domain knowledge gaps (e.g. no security experience for a frontend security role) are NOT dealbreakers — they belong in unmet.",
			},
		},
		Required: []string{"score", "met", "unmet", "dealbreakers"},
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
			const maxAttempts = 3
			for attempt := range maxAttempts {
				result, resp, err := scorePosting(gctx, client, resumeText, allowedLocations, p)
				if err != nil {
					var apierr *anthropic.Error
					if errors.As(err, &apierr) && apierr.StatusCode >= 400 && apierr.StatusCode < 500 && apierr.StatusCode != 429 {
						return fmt.Errorf("posting %d (%s): %w", p.ID, p.Title, err)
					}
					if attempt < maxAttempts-1 {
						log.Printf("  retry scoring posting %d (%s) (attempt %d): %v", p.ID, p.Title, attempt+1, err)
						continue
					}
					log.Printf("  error scoring posting %d (%s): %v", p.ID, p.Title, err)
					return nil
				}
				usage.add(resp)
				log.Printf("  [%.0f/10] %s", result.Score, p.Title)
				resultsCh <- *result
				return nil
			}
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
		Model:     "claude-sonnet-5-5",
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: fmt.Sprintf(`You are screening whether a candidate would realistically get an interview for this role.

Title: %s%s

Job Description:
%s

Scoring guide:
- 9-10: Meets all key requirements, no dealbreakers
- 7-8: Meets most requirements, minor gaps that would not block an interview
- 4-6: Mixed fit — some requirements met but notable gaps
- 1-3: Has a dealbreaker (location, experience level, missing core required skill)

Rules:
- A "nice to have" or "bonus" skill is never a dealbreaker
- Experience gap >2 years below the stated minimum is a dealbreaker
- "Or equivalent experience" means degree is not a dealbreaker
- Look for concrete evidence of fit, not keyword overlap%s`, p.Title, postingLocation, p.Description, locationConstraint)},
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
			OfAuto: &anthropic.ToolChoiceAutoParam{},
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
			PostingID:    p.ID,
			Score:        float64(out.Score),
			Met:          out.Met,
			Unmet:        out.Unmet,
			Dealbreakers: out.Dealbreakers,
		}, resp, nil
	}

	return nil, nil, fmt.Errorf("no score_posting tool use in response")
}
