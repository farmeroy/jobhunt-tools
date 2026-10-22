package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var AshbyBaseURL = "https://jobs.ashbyhq.com"

type Ashby struct {
	detailSem chan struct{}
}

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

	var wg sync.WaitGroup
	for i := range postings {
		if !MatchesTitle(postings[i].Title) || !MatchesLocation(postings[i].Location) {
			continue
		}
		wg.Add(1)
		go func(p *JobPosting) {
			defer wg.Done()
			a.detailSem <- struct{}{}
			defer func() { <-a.detailSem }()
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
	reqBody, err := json.Marshal(map[string]any{
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

	var lastErr error
	for attempt := range 4 {
		if attempt > 0 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		desc, retryAfter, err := a.doDetailRequest(ctx, client, reqBody)
		if err == nil {
			return desc, nil
		}
		lastErr = err

		if retryAfter == 0 {
			return "", err
		}
		if retryAfter > time.Duration(1<<(attempt))*time.Second {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(retryAfter):
			}
		}
	}
	return "", lastErr
}

// doDetailRequest makes one detail API call. Returns (description, retryAfter, error).
// retryAfter > 0 signals the caller should retry.
func (a *Ashby) doDetailRequest(ctx context.Context, client *http.Client, reqBody []byte) (string, time.Duration, error) {
	url := AshbyBaseURL + "/api/non-user-graphql?op=ApiJobPosting"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 || resp.StatusCode >= 500 {
		var retryAfter time.Duration
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil {
				retryAfter = time.Duration(secs) * time.Second
			}
		}
		return "", retryAfter + 1, fmt.Errorf("status %d", resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		return "", 0, fmt.Errorf("status %d", resp.StatusCode)
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
		return "", 0, fmt.Errorf("decode: %w", err)
	}
	if len(result.Errors) > 0 {
		return "", 0, fmt.Errorf("graphql: %s", result.Errors[0].Message)
	}

	return result.Data.JobPosting.DescriptionHtml, 0, nil
}
