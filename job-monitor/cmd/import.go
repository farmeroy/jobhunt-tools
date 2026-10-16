package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"jobhunt/tools/internal/atsutil"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <csv-file>",
	Short: "Import companies from a CSV file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()

		reader := csv.NewReader(f)
		header, err := reader.Read()
		if err != nil {
			return fmt.Errorf("read header: %w", err)
		}

		colIdx := map[string]int{}
		for i, h := range header {
			colIdx[h] = i
		}
		if _, ok := colIdx["company"]; !ok {
			return fmt.Errorf("CSV missing required 'company' column")
		}

		var imported, skippedNotFound, skippedDup int
		records, err := reader.ReadAll()
		if err != nil {
			return fmt.Errorf("read CSV: %w", err)
		}

		for _, record := range records {
			name := record[colIdx["company"]]
			ats := ""
			if idx, ok := colIdx["ats"]; ok {
				ats = record[idx]
			}

			if strings.HasPrefix(ats, "not-found") {
				skippedNotFound++
				continue
			}

			slugs := atsutil.Slugify(name)
			slug := ""
			if len(slugs) > 0 {
				slug = slugs[0]
			}

			_, err := db.AddCompany(name, slug, ats)
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") {
					skippedDup++
					fmt.Fprintf(cmd.ErrOrStderr(), "  skip (duplicate): %s\n", name)
					continue
				}
				return fmt.Errorf("add %s: %w", name, err)
			}
			imported++
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Imported %d, skipped %d (no ATS), skipped %d (duplicate)\n",
			imported, skippedNotFound, skippedDup)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
