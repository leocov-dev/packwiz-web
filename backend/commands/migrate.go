package commands

import (
	"github.com/spf13/cobra"
	"packwiz-web/internal/database"
	"packwiz-web/internal/log"
)

var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "migrate the database",
		Run: func(cmd *cobra.Command, _ []string) {
			if err := database.RunMigrations(); err != nil {
				log.Error("Migration failed:", err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}
