package cmd

import (
	"fmt"
	"strings"

	"jobhunt/tools/internal/atsutil"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <company>",
	Short: "Add a company and auto-detect its ATS",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		client := atsutil.NewHTTPClient()
		ats := atsutil.DetectATS(client, name)

		slugs := atsutil.Slugify(name)
		slug := ""
		if len(slugs) > 0 {
			slug = slugs[0]
		}

		_, err := db.AddCompany(name, slug, ats)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return fmt.Errorf("%s already exists", name)
			}
			return err
		}

		if ats != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Added %s (ATS: %s, slug: %s)\n", name, ats, slug)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Added %s (no ATS detected)\n", name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
