package commands

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"packwiz-web/internal/database"
	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/server"
	"packwiz-web/internal/services/packwiz_svc"
)

var (
	runMigrations bool
	runWorker     bool

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

			if runWorker {
				ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
				defer stop()

				db := database.GetClient()
				resolver := packwiz_svc.NewPackwizService(db, nil)
				client, err := jobs.NewClient(db, jobs.NewWorkers(resolver))
				if err != nil {
					log.Error("failed to create river client:", err)
					return
				}

				if err := client.Start(ctx); err != nil {
					log.Error("failed to start river client:", err)
					return
				}

				log.Info("worker started in-process")
			}

			server.Start()
		},
	}
)

func init() {
	startCmd.Flags().BoolVar(&runMigrations, "migrate", false, "run migrations before starting the server")
	startCmd.Flags().BoolVar(&runWorker, "worker", false, "run the background job worker in-process alongside the server")

	rootCmd.AddCommand(startCmd)
}
