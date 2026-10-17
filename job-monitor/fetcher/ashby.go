package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var AshbyBaseURL = "https://jobs.ashbyhq.com"

type Ashby struct{}

func (a *Ashby) Name() string { return "Ashby" }

func (a *Ashby) Fetch(ctx context.Context, client *http.Client, slug string) ([]JobPosting, error) {
	body, err := json.Marshal(map[string]any{
		"operationName": "ApiJobBoardWithTeams",
		"variables": map[string]string{
			"organizationHostedJobsPageName": slug,
		},
		"query": `query ApiJobBoardWithTeams($organizationHostedJobsPageName: String!) {
			jobBoard: jobBoardWithTeams(organizationHostedJobsPageName: $organizationHostedJobsPageName) {
				jobPostings { id title locationName }
			}
		}`,
	})
	if err != nil {
		return nil, err
	}

	url := AshbyBaseURL + "/api/non-user-graphql?op=ApiJobBoardWithTeams"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ashby: %s returned %d", slug, resp.StatusCode)
	}

	var result struct {
		Data struct {
			JobBoard struct {
				JobPostings []struct {
					ID           string `json:"id"`
					Title        string `json:"title"`
					LocationName string `json:"locationName"`
				} `json:"jobPostings"`
			} `json:"jobBoard"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ashby: decode: %w", err)
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("ashby: graphql: %s", result.Errors[0].Message)
	}

	jobs := result.Data.JobBoard.JobPostings
	postings := make([]JobPosting, len(jobs))
	for i, j := range jobs {
		postings[i] = JobPosting{
			ExternalID: j.ID,
			Title:      j.Title,
			URL:        fmt.Sprintf("%s/%s/%s", AshbyBaseURL, slug, j.ID),
			Location:   j.LocationName,
		}
	}
	return postings, nil
}
