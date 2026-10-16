package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <company>",
	Short: "Deactivate a company (soft delete)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if err := db.DeactivateCompany(name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deactivated %s\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
