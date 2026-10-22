package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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

	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	for i := range postings {
		if !MatchesTitle(postings[i].Title) || !MatchesLocation(postings[i].Location) {
			continue
		}
		wg.Add(1)
		go func(p *JobPosting) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			desc, err := a.fetchDescription(ctx, client, slug, p.ExternalID)
			if err != nil {
				log.Printf("  ashby: %s: description fetch failed for %s: %v", slug, p.Title, err)
				return
			}
			p.Description = desc
		}(&postings[i])
	}
	wg.Wait()

	return postings, nil
}

func (a *Ashby) fetchDescription(ctx context.Context, client *http.Client, slug, postingID string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"operationName": "ApiJobPosting",
		"variables": map[string]string{
			"organizationHostedJobsPageName": slug,
			"jobPostingId":                   postingID,
		},
		"query": `query ApiJobPosting($organizationHostedJobsPageName: String!, $jobPostingId: String!) {
			jobPosting(organizationHostedJobsPageName: $organizationHostedJobsPageName, jobPostingId: $jobPostingId) {
				descriptionHtml
			}
		}`,
	})
	if err != nil {
		return "", err
	}

	url := AshbyBaseURL + "/api/non-user-graphql?op=ApiJobPosting"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			JobPosting struct {
				DescriptionHtml string `json:"descriptionHtml"`
			} `json:"jobPosting"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if len(result.Errors) > 0 {
		return "", fmt.Errorf("graphql: %s", result.Errors[0].Message)
	}

	return result.Data.JobPosting.DescriptionHtml, nil
}
