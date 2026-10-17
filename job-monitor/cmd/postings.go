package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	postingsCompany string
	postingsNew     bool
)

var postingsCmd = &cobra.Command{
	Use:   "postings",
	Short: "List stored job postings",
	RunE: func(cmd *cobra.Command, args []string) error {
		postings, err := db.ListPostings(postingsCompany, postingsNew)
		if err != nil {
			return err
		}

		if len(postings) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No postings found.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COMPANY\tTITLE\tURL\tFIRST SEEN")
		for _, p := range postings {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				p.CompanyName, p.Title, p.URL, p.FirstSeen.Format("2006-01-02"))
		}
		return w.Flush()
	},
}

func init() {
	postingsCmd.Flags().StringVar(&postingsCompany, "company", "", "filter by company name")
	postingsCmd.Flags().BoolVar(&postingsNew, "new", false, "show only unfiltered (new) postings")
	rootCmd.AddCommand(postingsCmd)
}
