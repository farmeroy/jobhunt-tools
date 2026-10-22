package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
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

	respBody, err := ashbyPost(ctx, client, "ApiJobBoardWithTeams", body)
	if err != nil {
		return nil, fmt.Errorf("ashby: %s: %w", slug, err)
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
	if err := json.Unmarshal(respBody, &result); err != nil {
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

	for i := range postings {
		if !MatchesTitle(postings[i].Title) || !MatchesLocation(postings[i].Location) {
			continue
		}
		desc, err := a.fetchDescription(ctx, client, slug, postings[i].ExternalID)
		if err != nil {
			log.Printf("  ashby: %s: description fetch failed for %s: %v", slug, postings[i].Title, err)
			continue
		}
		postings[i].Description = desc
	}

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

	respBody, err := ashbyPost(ctx, client, "ApiJobPosting", reqBody)
	if err != nil {
		return "", err
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
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if len(result.Errors) > 0 {
		return "", fmt.Errorf("graphql: %s", result.Errors[0].Message)
	}

	return result.Data.JobPosting.DescriptionHtml, nil
}

// ashbyPost makes a POST to the Ashby GraphQL API with retry on 429/5xx.
func ashbyPost(ctx context.Context, client *http.Client, op string, body []byte) ([]byte, error) {
	url := AshbyBaseURL + "/api/non-user-graphql?op=" + op

	var lastErr error
	for attempt := range 4 {
		if attempt > 0 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil {
					extra := time.Duration(secs) * time.Second
					if extra > time.Duration(1<<attempt)*time.Second {
						select {
						case <-ctx.Done():
							return nil, ctx.Err()
						case <-time.After(extra):
						}
					}
				}
			}
			continue
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("status %d", resp.StatusCode)
		}

		return respBody, nil
	}
	return nil, lastErr
}
