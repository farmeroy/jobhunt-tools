package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"


	"jobhunt/tools/internal/atsutil"
	"jobhunt/tools/job-monitor/fetcher"
	"jobhunt/tools/job-monitor/filter"
	"jobhunt/tools/job-monitor/resume"
	"jobhunt/tools/job-monitor/store"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/spf13/cobra"
)

var (
	dryRun      bool
	resumePath  string
	reclassify  string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch job postings and classify titles for relevance",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		logDir := filepath.Join(filepath.Dir(dbPath), ".logs")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("create log directory: %w", err)
		}
		logFile, err := os.Create(filepath.Join(logDir, time.Now().UTC().Format("2006-01-02T150405Z")+".log"))
		if err != nil {
			return fmt.Errorf("create log file: %w", err)
		}
		defer logFile.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, logFile))

		// Load resume if provided (required for classification, not for dry-run)
		var resumeText string
		if resumePath != "" {
			var err error
			resumeText, err = resume.LoadText(resumePath)
			if err != nil {
				return fmt.Errorf("load resume: %w", err)
			}
		}

		// Validate API key availability for classification
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if !dryRun && resumePath != "" && apiKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY env var required for title classification")
		}

		// --- Fetch phase ---
		companies, err := db.ListActiveCompanies()
		if err != nil {
			return err
		}

		httpClient := atsutil.NewHTTPClient()
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

				postings, err := f.Fetch(ctx, httpClient, company.Slug)
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
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Checked %d companies: %d fetched, %d matched filter, %d new, %d skipped, %d errors\n",
			checked, totalFetched, totalFiltered, totalNew, skipped, errors)

		// --- Classification phase ---
		if resumeText == "" {
			log.Printf("  no --resume provided, skipping title classification")
			return nil
		}

		var since time.Time
		if reclassify != "" {
			d, err := parseDuration(reclassify)
			if err != nil {
				return fmt.Errorf("invalid --reclassify value %q: %w", reclassify, err)
			}
			since = time.Now().UTC().Add(-d)
			fmt.Fprintf(cmd.OutOrStdout(), "Reclassifying titles from postings since %s\n", since.Format("2006-01-02"))
		}

		titles, err := db.GetDistinctTitles(since)
		if err != nil {
			return fmt.Errorf("get titles: %w", err)
		}
		if len(titles) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No titles to classify")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Classifying %d unique titles...\n", len(titles))

		llmClient := anthropic.NewClient()
		classifications, usage, err := filter.FilterTitles(ctx, &llmClient, resumeText, titles)
		if err != nil {
			return fmt.Errorf("title classification: %w", err)
		}

		var relevant, irrelevant, postingsUpdated int
		for title, r := range classifications {
			n, err := db.SetTitleRelevanceByTitle(title, r)
			if err != nil {
				log.Printf("  error setting relevance for %q: %v", title, err)
				continue
			}
			postingsUpdated += int(n)
			if r {
				relevant++
			} else {
				irrelevant++
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Classified %d unique titles (%d relevant, %d irrelevant), updated %d postings\n",
			relevant+irrelevant, relevant, irrelevant, postingsUpdated)
		fmt.Fprintf(cmd.OutOrStdout(), "Token usage: %d input, %d output (%d API calls)\n",
			usage.InputTokens, usage.OutputTokens, usage.Calls)

		return nil
	},
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0, fmt.Errorf("expected number followed by d, w, or h (e.g. 7d, 2w, 24h)")
	}
	unit := s[len(s)-1]
	num, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return 0, fmt.Errorf("expected number followed by d, w, or h (e.g. 7d, 2w, 24h)")
	}
	switch unit {
	case 'h':
		return time.Duration(num) * time.Hour, nil
	case 'd':
		return time.Duration(num) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(num) * 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown unit %q, use d, w, or h", string(unit))
	}
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "fetch and filter but don't persist to database")
	runCmd.Flags().StringVar(&resumePath, "resume", "", "path to resume file (PDF, txt, or md) for title classification")
	runCmd.Flags().StringVar(&reclassify, "reclassify", "", "reclassify titles from postings within this period (e.g. 7d, 2w)")
	rootCmd.AddCommand(runCmd)
}
