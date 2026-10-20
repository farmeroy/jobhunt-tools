package notify

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"jobhunt/tools/job-monitor/store"
)

func WriteMatchesFile(path string, postings []store.Posting) error {
	if len(postings) == 0 {
		return nil
	}

	grouped := make(map[string][]store.Posting)
	for _, p := range postings {
		grouped[p.CompanyName] = append(grouped[p.CompanyName], p)
	}

	companies := make([]string, 0, len(grouped))
	for c := range grouped {
		companies = append(companies, c)
	}
	sort.Strings(companies)

	var b strings.Builder
	fmt.Fprintf(&b, "Job Monitor: %d new matches\n\n", len(postings))

	for _, company := range companies {
		fmt.Fprintf(&b, "=== %s ===\n\n", company)
		for _, p := range grouped[company] {
			fmt.Fprintf(&b, "  %s\n", p.Title)
			if p.ContentScore != nil {
				fmt.Fprintf(&b, "  Score: %.0f/10\n", *p.ContentScore)
			}
			if p.ContentReasoning != nil {
				fmt.Fprintf(&b, "  %s\n", *p.ContentReasoning)
			}
			fmt.Fprintf(&b, "  %s\n\n", p.URL)
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}
