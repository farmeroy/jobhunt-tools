package fetcher

import (
	"context"
	"net/http"
)

type JobPosting struct {
	ExternalID  string
	Title       string
	URL         string
	Description string
}

type Fetcher interface {
	Name() string
	Fetch(ctx context.Context, client *http.Client, slug string) ([]JobPosting, error)
}

var fetchers = map[string]Fetcher{
	"Greenhouse": &Greenhouse{},
	"Lever":      &Lever{},
	"Ashby":      &Ashby{},
}

func ForProvider(ats string) (Fetcher, bool) {
	f, ok := fetchers[ats]
	return f, ok
}
