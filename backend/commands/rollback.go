package commands

import (
	"github.com/spf13/cobra"
	"packwiz-web/internal/database"
	"packwiz-web/internal/log"
	"strconv"
)

var (
	rollbackCmd = &cobra.Command{
		Use:   "rollback",
		Short: "rollback the database N steps",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			steps, err := strconv.Atoi(args[0])
			if err != nil {
				log.Error("invalid steps, must be integer", err)
				return
			}
			if err := database.RollbackMigration(steps); err != nil {
				log.Error("Migration rollback failed:", err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
