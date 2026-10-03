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
	"packwiz-web/internal/server"
	"packwiz-web/internal/services/audit_svc"
	"packwiz-web/internal/services/pack_access_svc"
	"packwiz-web/internal/services/packwiz_svc"
)

var (
	runMigrations bool
	runWorker     bool

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the server",
		Run: func(cmd *cobra.Command, args []string) {
			// SIGINT/SIGTERM must stop the whole process, not just the worker:
			// once a handler is installed the default "exit on signal" is gone,
			// so the HTTP server has to watch this context too.
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			if runMigrations {
				if err := database.RunMigrations(); err != nil {
					log.Error("Migration failed:", err)
					return
				}
			}

			database.UpsertDefaultAdminUser()

			if runWorker {
				db := database.GetClient()
				resolver := packwiz_svc.NewPackwizService(db, nil)
				client, err := jobs.NewClient(db, jobs.NewWorkers(resolver, resolver, audit_svc.NewAuditService(db), pack_access_svc.NewPackAccessService(db)))
				if err != nil {
					log.Error("failed to create river client:", err)
					return
				}

				if err := client.Start(ctx); err != nil {
					log.Error("failed to start river client:", err)
					return
				}

				log.Info("worker started in-process")

				defer func() {
					stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					if err := client.Stop(stopCtx); err != nil {
						log.Error("error stopping river client:", err)
					}
					log.Info("worker stopped")
				}()
			}

			server.Start(ctx)
		},
	}
)

func init() {
	startCmd.Flags().BoolVar(&runMigrations, "migrate", false, "run migrations before starting the server")
	startCmd.Flags().BoolVar(&runWorker, "worker", false, "run the background job worker in-process alongside the server")

	rootCmd.AddCommand(startCmd)
}
