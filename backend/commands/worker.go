package commands

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"packwiz-web/internal/database"
	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/packwiz_svc"
)

var (
	workerCmd = &cobra.Command{
		Use:   "worker",
		Short: "Run the background job worker process",
		Run: func(cmd *cobra.Command, args []string) {
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

			log.Info("worker started")

			<-ctx.Done()
			log.Info("shutdown signal received, stopping worker...")

			stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := client.Stop(stopCtx); err != nil {
				log.Error("error stopping river client:", err)
			}

			log.Info("worker stopped")
		},
	}
)

func init() {
	rootCmd.AddCommand(workerCmd)
}
