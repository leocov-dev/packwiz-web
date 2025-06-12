package commands

import (
	"github.com/spf13/cobra"
	"packwiz-web/internal/database"
	"packwiz-web/internal/log"
	"packwiz-web/internal/server"
)

var (
	runMigrations bool

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the server",
		Run: func(cmd *cobra.Command, args []string) {

			if runMigrations {
				if err := database.RunMigrations(); err != nil {
					log.Error("Migration failed:", err)
					return
				}
			}

			database.UpsertDefaultAdminUser()

			server.Start()
		},
	}
)

func init() {
	startCmd.Flags().BoolVar(&runMigrations, "migrate", false, "run migrations before starting the server")

	rootCmd.AddCommand(startCmd)
}
