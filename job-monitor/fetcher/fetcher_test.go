package fetcher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestGreenhouseFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"jobs": []map[string]any{
				{
					"id":           1234,
					"title":        "Software Engineer",
					"absolute_url": "https://boards.greenhouse.io/acme/jobs/1234",
					"content":      "<p>Build things</p>",
				},
				{
					"id":           5678,
					"title":        "Product Manager",
					"absolute_url": "https://boards.greenhouse.io/acme/jobs/5678",
					"content":      "<p>Manage things</p>",
				},
			},
		})
	}))
	defer srv.Close()

	old := GreenhouseBaseURL
	GreenhouseBaseURL = srv.URL
	defer func() { GreenhouseBaseURL = old }()

	g := &Greenhouse{}
	postings, err := g.Fetch(context.Background(), srv.Client(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(postings) != 2 {
		t.Fatalf("expected 2 postings, got %d", len(postings))
	}
	if postings[0].ExternalID != "1234" {
		t.Errorf("expected id 1234, got %s", postings[0].ExternalID)
	}
	if postings[0].Title != "Software Engineer" {
		t.Errorf("expected Software Engineer, got %s", postings[0].Title)
	}
	if postings[0].Description != "<p>Build things</p>" {
		t.Errorf("expected description, got %s", postings[0].Description)
	}
}

func TestGreenhouseEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"jobs": []any{}})
	}))
	defer srv.Close()

	old := GreenhouseBaseURL
	GreenhouseBaseURL = srv.URL
	defer func() { GreenhouseBaseURL = old }()

	postings, err := (&Greenhouse{}).Fetch(context.Background(), srv.Client(), "empty")
	if err != nil {
		t.Fatal(err)
	}
	if len(postings) != 0 {
		t.Fatalf("expected 0 postings, got %d", len(postings))
	}
}

func TestGreenhouseNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()

	old := GreenhouseBaseURL
	GreenhouseBaseURL = srv.URL
	defer func() { GreenhouseBaseURL = old }()

	_, err := (&Greenhouse{}).Fetch(context.Background(), srv.Client(), "nope")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestLeverFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":               "abc-123",
				"text":             "Backend Engineer",
				"hostedUrl":        "https://jobs.lever.co/acme/abc-123",
				"descriptionPlain": "Write backend code",
			},
		})
	}))
	defer srv.Close()

	old := LeverBaseURL
	LeverBaseURL = srv.URL
	defer func() { LeverBaseURL = old }()

	postings, err := (&Lever{}).Fetch(context.Background(), srv.Client(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(postings) != 1 {
		t.Fatalf("expected 1 posting, got %d", len(postings))
	}
	if postings[0].ExternalID != "abc-123" {
		t.Errorf("expected abc-123, got %s", postings[0].ExternalID)
	}
	if postings[0].Title != "Backend Engineer" {
		t.Errorf("expected Backend Engineer, got %s", postings[0].Title)
	}
	if postings[0].Description != "Write backend code" {
		t.Errorf("expected description, got %s", postings[0].Description)
	}
}

func TestLeverNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":false,"error":"Document not found"}`))
	}))
	defer srv.Close()

	old := LeverBaseURL
	LeverBaseURL = srv.URL
	defer func() { LeverBaseURL = old }()

	_, err := (&Lever{}).Fetch(context.Background(), srv.Client(), "nope")
	if err == nil {
		t.Fatal("expected error for invalid response")
	}
}

func TestAshbyFetch(t *testing.T) {
	descriptions := map[string]string{
		"uuid-1": "<p>Build full stack apps</p>",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OperationName string         `json:"operationName"`
			Variables     map[string]string `json:"variables"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		switch req.OperationName {
		case "ApiJobBoardWithTeams":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"jobBoard": map[string]any{
						"jobPostings": []map[string]any{
							{
								"id":           "uuid-1",
								"title":        "Full Stack Engineer",
								"locationName": "Remote",
							},
							{
								"id":           "uuid-2",
								"title":        "Data Scientist",
								"locationName": "NYC",
							},
						},
					},
				},
			})
		case "ApiJobPosting":
			id := req.Variables["jobPostingId"]
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"jobPosting": map[string]any{
						"descriptionHtml": descriptions[id],
					},
				},
			})
		}
	}))
	defer srv.Close()

	old := AshbyBaseURL
	AshbyBaseURL = srv.URL
	defer func() { AshbyBaseURL = old }()

	postings, err := (&Ashby{}).Fetch(context.Background(), srv.Client(), "testco")
	if err != nil {
		t.Fatal(err)
	}
	if len(postings) != 2 {
		t.Fatalf("expected 2 postings, got %d", len(postings))
	}
	if postings[0].ExternalID != "uuid-1" {
		t.Errorf("expected uuid-1, got %s", postings[0].ExternalID)
	}
	if postings[0].Title != "Full Stack Engineer" {
		t.Errorf("expected Full Stack Engineer, got %s", postings[0].Title)
	}
	if postings[0].URL != srv.URL+"/testco/uuid-1" {
		t.Errorf("unexpected URL: %s", postings[0].URL)
	}
	if postings[0].Description != "<p>Build full stack apps</p>" {
		t.Errorf("expected description for matched posting, got %q", postings[0].Description)
	}
	if postings[1].Description != "" {
		t.Errorf("expected empty description for non-engineer posting, got %q", postings[1].Description)
	}
}

func TestAshbyGraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{
				{"message": "something broke"},
			},
		})
	}))
	defer srv.Close()

	old := AshbyBaseURL
	AshbyBaseURL = srv.URL
	defer func() { AshbyBaseURL = old }()

	_, err := (&Ashby{}).Fetch(context.Background(), srv.Client(), "bad")
	if err == nil {
		t.Fatal("expected error for graphql error response")
	}
}

func TestAshbyFetchRetryDetail(t *testing.T) {
	var detailCalls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OperationName string            `json:"operationName"`
			Variables     map[string]string `json:"variables"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		switch req.OperationName {
		case "ApiJobBoardWithTeams":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"jobBoard": map[string]any{
						"jobPostings": []map[string]any{
							{"id": "uuid-1", "title": "Software Engineer", "locationName": "Remote"},
						},
					},
				},
			})
		case "ApiJobPosting":
			n := detailCalls.Add(1)
			if n <= 2 {
				w.WriteHeader(429)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"jobPosting": map[string]any{
						"descriptionHtml": "<p>Got it on retry</p>",
					},
				},
			})
		}
	}))
	defer srv.Close()

	old := AshbyBaseURL
	AshbyBaseURL = srv.URL
	defer func() { AshbyBaseURL = old }()

	postings, err := (&Ashby{}).Fetch(context.Background(), srv.Client(), "testco")
	if err != nil {
		t.Fatal(err)
	}
	if postings[0].Description != "<p>Got it on retry</p>" {
		t.Errorf("expected description after retry, got %q", postings[0].Description)
	}
	if detailCalls.Load() != 3 {
		t.Errorf("expected 3 detail calls (2 retries + 1 success), got %d", detailCalls.Load())
	}
}

func TestAshbyFetchRetryListing(t *testing.T) {
	var listCalls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OperationName string            `json:"operationName"`
			Variables     map[string]string `json:"variables"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		switch req.OperationName {
		case "ApiJobBoardWithTeams":
			n := listCalls.Add(1)
			if n <= 2 {
				w.WriteHeader(429)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"jobBoard": map[string]any{
						"jobPostings": []map[string]any{
							{"id": "uuid-1", "title": "Software Engineer", "locationName": "Remote"},
						},
					},
				},
			})
		case "ApiJobPosting":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"jobPosting": map[string]any{
						"descriptionHtml": "<p>Description</p>",
					},
				},
			})
		}
	}))
	defer srv.Close()

	old := AshbyBaseURL
	AshbyBaseURL = srv.URL
	defer func() { AshbyBaseURL = old }()

	postings, err := (&Ashby{}).Fetch(context.Background(), srv.Client(), "testco")
	if err != nil {
		t.Fatal(err)
	}
	if len(postings) != 1 {
		t.Fatalf("expected 1 posting after listing retry, got %d", len(postings))
	}
	if listCalls.Load() != 3 {
		t.Errorf("expected 3 listing calls (2 retries + 1 success), got %d", listCalls.Load())
	}
}

func TestForProvider(t *testing.T) {
	for _, name := range []string{"Greenhouse", "Lever", "Ashby"} {
		f, ok := ForProvider(name)
		if !ok {
			t.Errorf("expected fetcher for %s", name)
		}
		if f.Name() != name {
			t.Errorf("expected name %s, got %s", name, f.Name())
		}
	}

	_, ok := ForProvider("Workday")
	if ok {
		t.Error("expected no fetcher for Workday")
	}
}
