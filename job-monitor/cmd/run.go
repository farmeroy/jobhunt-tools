package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"

	"jobhunt/tools/internal/atsutil"
	"jobhunt/tools/job-monitor/fetcher"
	"jobhunt/tools/job-monitor/store"

	"github.com/spf13/cobra"
)

var dryRun bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch job postings for all active companies",
	RunE: func(cmd *cobra.Command, args []string) error {
		companies, err := db.ListActiveCompanies()
		if err != nil {
			return err
		}

		client := atsutil.NewHTTPClient()
		ctx := context.Background()
		sem := make(chan struct{}, 5)

		type result struct {
			company  string
			fetched  int
			filtered int
			newCount int
			err      error
		}

		results := make([]result, len(companies))
		var wg sync.WaitGroup
		var skipped int

		for i, c := range companies {
			f, ok := fetcher.ForProvider(c.ATS)
			if !ok {
				log.Printf("  skip %s (no fetcher for %q)", c.Name, c.ATS)
				skipped++
				continue
			}

			wg.Add(1)
			go func(idx int, company store.Company, f fetcher.Fetcher) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				postings, err := f.Fetch(ctx, client, company.Slug)
				if err != nil {
					log.Printf("  error %s: %v", company.Name, err)
					results[idx] = result{company: company.Name, err: err}
					return
				}

				matched := fetcher.PreFilter(postings)

				if dryRun {
					log.Printf("  [dry-run] %s: %d fetched, %d matched filter", company.Name, len(postings), len(matched))
					results[idx] = result{company: company.Name, fetched: len(postings), filtered: len(matched)}
					return
				}

				storePostings := make([]store.Posting, len(matched))
				for j, p := range matched {
					storePostings[j] = store.Posting{
						CompanyID:   company.ID,
						ExternalID:  p.ExternalID,
						Title:       p.Title,
						URL:         p.URL,
						Location:    p.Location,
						Description: p.Description,
					}
				}

				newCount, err := db.UpsertPostings(company.ID, storePostings)
				if err != nil {
					log.Printf("  error %s (upsert): %v", company.Name, err)
					results[idx] = result{company: company.Name, err: err}
					return
				}

				// Use all fetched IDs for staleness — not just filtered ones
				allIDs := make([]string, len(postings))
				for j, p := range postings {
					allIDs[j] = p.ExternalID
				}
				if err := db.MarkStale(company.ID, allIDs); err != nil {
					log.Printf("  error %s (mark stale): %v", company.Name, err)
				}

				log.Printf("  %s: %d fetched, %d matched, %d new", company.Name, len(postings), len(matched), newCount)
				results[idx] = result{company: company.Name, fetched: len(postings), filtered: len(matched), newCount: newCount}
			}(i, c, f)
		}
		wg.Wait()

		var checked, totalFetched, totalFiltered, totalNew, errors int
		for _, r := range results {
			if r.company == "" {
				continue
			}
			if r.err != nil {
				errors++
				continue
			}
			checked++
			totalFetched += r.fetched
			totalFiltered += r.filtered
			totalNew += r.newCount
		}

		if dryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] Checked %d companies: %d fetched, %d matched filter, %d skipped, %d errors\n",
				checked, totalFetched, totalFiltered, skipped, errors)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Checked %d companies: %d fetched, %d matched filter, %d new, %d skipped, %d errors\n",
				checked, totalFetched, totalFiltered, totalNew, skipped, errors)
		}
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "fetch and filter but don't persist to database")
	rootCmd.AddCommand(runCmd)
}
