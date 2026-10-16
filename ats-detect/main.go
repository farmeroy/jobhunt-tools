package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"jobhunt/tools/internal/atsutil"
)

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

	client := atsutil.NewHTTPClient()
	results := make(chan result, len(toCheck))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, idx := range toCheck {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			company := rows[i].Company
			ats := atsutil.DetectATS(client, company)
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
