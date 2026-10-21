package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

const chunkSize = 100

type titleResult struct {
	Index    int  `json:"index"`
	Relevant bool `json:"relevant"`
}

type titleResults struct {
	Results []titleResult `json:"results"`
}

type Usage struct {
	InputTokens  int64
	OutputTokens int64
	Calls        int64
}

func (u *Usage) add(resp *anthropic.Message) {
	atomic.AddInt64(&u.InputTokens, resp.Usage.InputTokens)
	atomic.AddInt64(&u.OutputTokens, resp.Usage.OutputTokens)
	atomic.AddInt64(&u.Calls, 1)
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
						"index":    map[string]any{"type": "integer", "description": "The index number of the title from the list"},
						"relevant": map[string]any{"type": "boolean", "description": "true if the title is relevant to the resume"},
					},
					"required": []string{"index", "relevant"},
				},
			},
		},
		Required: []string{"results"},
	},
}

func FilterTitles(ctx context.Context, client *anthropic.Client, resumeText string, titles []string) (map[string]bool, *Usage, error) {
	if len(titles) == 0 {
		return nil, &Usage{}, nil
	}

	usage := &Usage{}

	log.Printf("  inferring target titles from resume...")
	targetTitles, resp, err := inferTargetTitles(ctx, client, resumeText)
	if err != nil {
		return nil, nil, fmt.Errorf("infer target titles: %w", err)
	}
	usage.add(resp)
	log.Printf("  target titles:\n%s", targetTitles)

	results := make(map[string]bool, len(titles))

	for i := 0; i < len(titles); i += chunkSize {
		end := i + chunkSize
		if end > len(titles) {
			end = len(titles)
		}
		chunk := titles[i:end]

		chunkResults, chunkResp, err := classifyChunk(ctx, client, resumeText, targetTitles, chunk, i)
		if err != nil {
			return nil, nil, fmt.Errorf("classify chunk %d-%d: %w", i, end, err)
		}
		usage.add(chunkResp)
		for title, relevant := range chunkResults {
			results[title] = relevant
			if relevant {
				log.Printf("    [relevant] %s", title)
			} else {
				log.Printf("    [skip]     %s", title)
			}
		}

		missing := findMissing(chunk, chunkResults)
		if len(missing) > 0 {
			log.Printf("  retry: model skipped %d titles, retrying", len(missing))
			retryResults, retryResp, err := classifyChunk(ctx, client, resumeText, targetTitles, missing, 0)
			if err != nil {
				log.Printf("  warning: retry failed: %v", err)
			} else {
				usage.add(retryResp)
				for title, relevant := range retryResults {
					results[title] = relevant
					if relevant {
						log.Printf("    [relevant] %s", title)
					} else {
						log.Printf("    [skip]     %s", title)
					}
				}
				stillMissing := findMissing(missing, retryResults)
				if len(stillMissing) > 0 {
					log.Printf("  warning: %d titles still unclassified after retry", len(stillMissing))
				}
			}
		}
	}

	return results, usage, nil
}

func findMissing(titles []string, results map[string]bool) []string {
	var missing []string
	for _, t := range titles {
		if _, ok := results[t]; !ok {
			missing = append(missing, t)
		}
	}
	return missing
}

func inferTargetTitles(ctx context.Context, client *anthropic.Client, resumeText string) (string, *anthropic.Message, error) {
	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			{
				Role: "user",
				Content: []anthropic.ContentBlockParamUnion{
					{OfText: &anthropic.TextBlockParam{Text: fmt.Sprintf("Given this resume, list the job titles this candidate should be targeting. Consider their skills, seniority level, and domain expertise. Be conservative with seniority — prefer downleveling over stretch roles. If the resume suggests Senior level, include Senior and Mid-level titles but exclude Staff, Principal, or Director titles. Only include levels the candidate clearly qualifies for, not aspirational ones.\n\nResume:\n\n%s", resumeText)}},
				},
			},
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("infer target titles: %w", err)
	}

	for _, block := range resp.Content {
		if block.Type == "text" {
			return block.Text, resp, nil
		}
	}
	return "", nil, fmt.Errorf("no text in response")
}

func classifyChunk(ctx context.Context, client *anthropic.Client, resumeText string, targetTitles string, titles []string, offset int) (map[string]bool, *anthropic.Message, error) {
	var listing strings.Builder
	for i, t := range titles {
		fmt.Fprintf(&listing, "%d. %s\n", offset+i, t)
	}

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: fmt.Sprintf("You are a job relevance classifier. You have already analyzed the candidate's resume and determined they should be targeting these roles:\n\n%s\n\nClassify each job title in the list as relevant or irrelevant based on whether it matches the target roles above. A title is relevant only if it aligns with the candidate's target seniority, skills, and domain. Classify every title in the list. Do not skip any.", targetTitles)},
		},
		Messages: []anthropic.MessageParam{
			{
				Role: "user",
				Content: []anthropic.ContentBlockParamUnion{
					{OfText: &anthropic.TextBlockParam{Text: fmt.Sprintf("Classify each of these job titles:\n\n%s", listing.String())}},
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
		return nil, nil, fmt.Errorf("anthropic api: %w", err)
	}

	results, err := parseToolUseResponse(resp, titles, offset)
	return results, resp, err
}

func parseToolUseResponse(resp *anthropic.Message, titles []string, offset int) (map[string]bool, error) {
	for _, block := range resp.Content {
		if block.Type != "tool_use" || block.Name != "classify_titles" {
			continue
		}

		var tr titleResults
		if err := json.Unmarshal(block.Input, &tr); err != nil {
			return nil, fmt.Errorf("parse tool response: %w", err)
		}

		results := make(map[string]bool, len(tr.Results))
		for _, r := range tr.Results {
			idx := r.Index - offset
			if idx < 0 || idx >= len(titles) {
				log.Printf("  warning: model returned out-of-range index %d", r.Index)
				continue
			}
			results[titles[idx]] = r.Relevant
		}
		return results, nil
	}

	return nil, fmt.Errorf("no classify_titles tool use in response")
}
