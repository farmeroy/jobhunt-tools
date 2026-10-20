package cmd

import (
	"fmt"
	"text/tabwriter"

	"jobhunt/tools/job-monitor/store"

	"github.com/spf13/cobra"
)

var (
	postingsCompany   string
	postingsNew       bool
	postingsMinScore  float64
)

var postingsCmd = &cobra.Command{
	Use:   "postings",
	Short: "List stored job postings",
	RunE: func(cmd *cobra.Command, args []string) error {
		var postings []store.Posting
		var err error

		if postingsMinScore > 0 {
			postings, err = db.GetUnnotifiedMatches(postingsMinScore)
		} else {
			postings, err = db.ListPostings(postingsCompany, postingsNew)
		}
		if err != nil {
			return err
		}

		if len(postings) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No postings found.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		if postingsMinScore > 0 {
			fmt.Fprintln(w, "COMPANY\tTITLE\tSCORE\tREASONING\tURL")
			for _, p := range postings {
				score := "-"
				if p.ContentScore != nil {
					score = fmt.Sprintf("%.0f", *p.ContentScore)
				}
				reasoning := ""
				if p.ContentReasoning != nil {
					reasoning = *p.ContentReasoning
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					p.CompanyName, p.Title, score, reasoning, p.URL)
			}
		} else {
			fmt.Fprintln(w, "COMPANY\tTITLE\tRELEVANT\tURL\tFIRST SEEN")
			for _, p := range postings {
				rel := "-"
				if p.TitleRelevant != nil {
					if *p.TitleRelevant {
						rel = "yes"
					} else {
						rel = "no"
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					p.CompanyName, p.Title, rel, p.URL, p.FirstSeen.Format("2006-01-02"))
			}
		}
		return w.Flush()
	},
}

func init() {
	postingsCmd.Flags().StringVar(&postingsCompany, "company", "", "filter by company name")
	postingsCmd.Flags().BoolVar(&postingsNew, "new", false, "show only unfiltered (new) postings")
	postingsCmd.Flags().Float64Var(&postingsMinScore, "min-score", 0, "show only postings scored at or above this threshold")
	rootCmd.AddCommand(postingsCmd)
}
