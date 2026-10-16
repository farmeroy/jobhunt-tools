package atsutil

import (
	"io"
	"net/http"
	"time"
)

type Provider struct {
	Name    string
	BaseURL string
	// MinSize is the minimum response body size to consider a board "real".
	// 0 means rely on HTTP status code only.
	MinSize int64
}

var Providers = []Provider{
	{Name: "Greenhouse", BaseURL: "https://boards.greenhouse.io/", MinSize: 0},
	{Name: "Lever", BaseURL: "https://jobs.lever.co/", MinSize: 0},
	{Name: "Ashby", BaseURL: "https://jobs.ashbyhq.com/", MinSize: 15000},
}

func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func CheckProvider(client *http.Client, provider Provider, slug string) bool {
	url := provider.BaseURL + slug
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	if provider.MinSize > 0 {
		buf := make([]byte, provider.MinSize+1)
		n, _ := io.ReadFull(resp.Body, buf)
		return int64(n) > provider.MinSize
	}

	return true
}

func DetectATS(client *http.Client, company string) string {
	slugs := Slugify(company)

	for _, provider := range Providers {
		for _, slug := range slugs {
			if CheckProvider(client, provider, slug) {
				return provider.Name
			}
		}
	}
	return ""
}
