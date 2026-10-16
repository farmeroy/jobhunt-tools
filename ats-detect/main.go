package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

type atsProvider struct {
	Name    string
	BaseURL string
	// MinSize is the minimum response body size to consider a board "real".
	// 0 means rely on HTTP status code only.
	MinSize int64
}

var providers = []atsProvider{
	{Name: "Greenhouse", BaseURL: "https://boards.greenhouse.io/", MinSize: 0},
	{Name: "Lever", BaseURL: "https://jobs.lever.co/", MinSize: 0},
	{Name: "Ashby", BaseURL: "https://jobs.ashbyhq.com/", MinSize: 15000},
}

var client = &http.Client{
	Timeout: 8 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return http.ErrUseLastResponse
		}
		return nil
	},
}

func slugify(company string) []string {
	lower := strings.ToLower(company)

	// Strip non-alphanumeric except spaces and hyphens
	clean := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			return r
		}
		return -1
	}, lower)

	// Collapse whitespace
	fields := strings.Fields(clean)
	joined := strings.Join(fields, "")       // "airbnb"
	dashed := strings.Join(fields, "-")       // "air-bnb" (only different for multi-word)

	slugs := []string{joined}
	if dashed != joined {
		slugs = append(slugs, dashed)
	}

	// Try without common suffixes
	for _, suffix := range []string{"ai", "io", "co", "inc", "health"} {
		for _, s := range []string{joined, dashed} {
			trimmed := strings.TrimSuffix(s, suffix)
			trimmed = strings.TrimSuffix(trimmed, "-")
			if trimmed != "" && trimmed != s {
				slugs = append(slugs, trimmed)
			}
		}
	}

	// Deduplicate
	seen := map[string]bool{}
	var unique []string
	for _, s := range slugs {
		if !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}
	return unique
}

func checkProvider(provider atsProvider, slug string) bool {
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
		// Read up to MinSize+1 to check if body exceeds threshold
		buf := make([]byte, provider.MinSize+1)
		n, _ := io.ReadFull(resp.Body, buf)
		return int64(n) > provider.MinSize
	}

	return true
}

func detectATS(company string) string {
	slugs := slugify(company)

	for _, provider := range providers {
		for _, slug := range slugs {
			if checkProvider(provider, slug) {
				return provider.Name
			}
		}
	}
	return ""
}

type row struct {
	Company        string
	ATS            string
	InterviewCount string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: ats-detect <companies.csv>")
		os.Exit(1)
	}

	csvPath := os.Args[1]

	f, err := os.Open(csvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %v\n", csvPath, err)
		os.Exit(1)
	}

	reader := csv.NewReader(f)
	header, err := reader.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading header: %v\n", err)
		os.Exit(1)
	}

	// Find column indices
	colIdx := map[string]int{}
	for i, h := range header {
		colIdx[h] = i
	}

	var rows []row
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		rows = append(rows, row{
			Company:        record[colIdx["company"]],
			ATS:            record[colIdx["ats"]],
			InterviewCount: record[colIdx["interview_count"]],
		})
	}
	f.Close()

	// Find rows that need checking (ats is empty)
	type result struct {
		idx int
		ats string
	}

	var toCheck []int
	for i, r := range rows {
		if r.ATS == "" {
			toCheck = append(toCheck, i)
		}
	}

	fmt.Printf("Total: %d companies, %d already have ATS, %d to check\n",
		len(rows), len(rows)-len(toCheck), len(toCheck))

	results := make(chan result, len(toCheck))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // concurrency limit

	for _, idx := range toCheck {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			company := rows[i].Company
			ats := detectATS(company)
			if ats != "" {
				fmt.Printf("  ✓ %s → %s\n", company, ats)
			} else {
				fmt.Printf("  · %s → not found\n", company)
			}
			results <- result{idx: i, ats: ats}
		}(idx)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	today := time.Now().Format("2006-01-02")
	for r := range results {
		if r.ats != "" {
			rows[r.idx].ATS = r.ats
		} else {
			rows[r.idx].ATS = "not-found-" + today
		}
	}

	// Write back
	out, err := os.Create(csvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", csvPath, err)
		os.Exit(1)
	}
	defer out.Close()

	writer := csv.NewWriter(out)
	writer.Write(header)
	for _, r := range rows {
		writer.Write([]string{r.Company, r.ATS, r.InterviewCount})
	}
	writer.Flush()

	// Summary
	counts := map[string]int{}
	for _, r := range rows {
		if strings.HasPrefix(r.ATS, "not-found") {
			counts["not-found"]++
		} else {
			counts[r.ATS]++
		}
	}
	fmt.Println("\n=== Summary ===")
	for ats, count := range counts {
		fmt.Printf("  %s: %d\n", ats, count)
	}
}
