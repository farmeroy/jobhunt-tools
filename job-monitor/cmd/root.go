package cmd

import (
	"fmt"
	"os"

	"jobhunt/tools/job-monitor/store"

	"github.com/spf13/cobra"
)

var (
	dbPath string
	db     *store.Store
)

var rootCmd = &cobra.Command{
	Use:   "job-monitor",
	Short: "Monitor companies for new job postings",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		db, err = store.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if db != nil {
			return db.Close()
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "job-monitor.db", "path to SQLite database")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
