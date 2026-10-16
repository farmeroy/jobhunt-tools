package cmd

import (
	"fmt"
	"text/tabwriter"

	"jobhunt/tools/job-monitor/store"

	"github.com/spf13/cobra"
)

var listAll bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List monitored companies",
	RunE: func(cmd *cobra.Command, args []string) error {
		var companies []store.Company
		var err error

		if listAll {
			companies, err = db.ListAllCompanies()
		} else {
			companies, err = db.ListActiveCompanies()
		}
		if err != nil {
			return err
		}

		if len(companies) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No companies found.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		if listAll {
			fmt.Fprintln(w, "NAME\tATS\tSLUG\tACTIVE")
			for _, c := range companies {
				activeStr := "yes"
				if !c.Active {
					activeStr = "no"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.ATS, c.Slug, activeStr)
			}
		} else {
			fmt.Fprintln(w, "NAME\tATS\tSLUG")
			for _, c := range companies {
				fmt.Fprintf(w, "%s\t%s\t%s\n", c.Name, c.ATS, c.Slug)
			}
		}
		return w.Flush()
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "include inactive companies")
	rootCmd.AddCommand(listCmd)
}
