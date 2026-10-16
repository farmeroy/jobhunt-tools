package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

var GreenhouseBaseURL = "https://boards-api.greenhouse.io/v1/boards"

type Greenhouse struct{}

func (g *Greenhouse) Name() string { return "Greenhouse" }

func (g *Greenhouse) Fetch(ctx context.Context, client *http.Client, slug string) ([]JobPosting, error) {
	url := fmt.Sprintf("%s/%s/jobs?content=true", GreenhouseBaseURL, slug)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("greenhouse: %s returned %d", slug, resp.StatusCode)
	}

	var result struct {
		Jobs []struct {
			ID          int64  `json:"id"`
			Title       string `json:"title"`
			AbsoluteURL string `json:"absolute_url"`
			Content     string `json:"content"`
		} `json:"jobs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("greenhouse: decode: %w", err)
	}

	postings := make([]JobPosting, len(result.Jobs))
	for i, j := range result.Jobs {
		postings[i] = JobPosting{
			ExternalID:  strconv.FormatInt(j.ID, 10),
			Title:       j.Title,
			URL:         j.AbsoluteURL,
			Description: j.Content,
		}
	}
	return postings, nil
}
