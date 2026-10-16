package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var LeverBaseURL = "https://api.lever.co/v0/postings"

type Lever struct{}

func (l *Lever) Name() string { return "Lever" }

func (l *Lever) Fetch(ctx context.Context, client *http.Client, slug string) ([]JobPosting, error) {
	url := fmt.Sprintf("%s/%s?mode=json", LeverBaseURL, slug)
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
		return nil, fmt.Errorf("lever: %s returned %d", slug, resp.StatusCode)
	}

	var jobs []struct {
		ID               string `json:"id"`
		Text             string `json:"text"`
		HostedURL        string `json:"hostedUrl"`
		DescriptionPlain string `json:"descriptionPlain"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return nil, fmt.Errorf("lever: decode: %w", err)
	}

	postings := make([]JobPosting, len(jobs))
	for i, j := range jobs {
		postings[i] = JobPosting{
			ExternalID:  j.ID,
			Title:       j.Text,
			URL:         j.HostedURL,
			Description: j.DescriptionPlain,
		}
	}
	return postings, nil
}
