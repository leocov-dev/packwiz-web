package commands

import (
	"github.com/spf13/cobra"
	"packwiz-web/internal/database"
	"packwiz-web/internal/server"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the server",
		Run: func(cmd *cobra.Command, args []string) {
			database.InitDb()
			database.CreateDefaultAdminUser()

			server.Start()
		},
	}
)

func init() {
	rootCmd.AddCommand(startCmd)
}
