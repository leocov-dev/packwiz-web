package commands

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"packwiz-web/internal/config"
	"packwiz-web/internal/database"
	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/audit_svc"
	"packwiz-web/internal/services/compat_svc"
	"packwiz-web/internal/services/pack_access_svc"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/services/webhook_svc"
)

var (
	workerCmd = &cobra.Command{
		Use:   "worker",
		Short: "Run the background job worker process",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			db := database.GetClient()
			compat := compat_svc.NewCompatService(db, nil, nil)
			webhook := webhook_svc.NewService(db, nil, config.C.SessionSecret)
			// jars of a previously killed worker would otherwise stay on disk
			if err := compat.PurgeTempFiles(); err != nil {
				log.Error("failed to purge jar temp files:", err)
			}
			resolver := packwiz_svc.NewPackwizService(db, nil, webhook.Chain(compat))
			client, err := jobs.NewClient(db, jobs.NewWorkers(resolver, resolver, resolver, audit_svc.NewAuditService(db), pack_access_svc.NewPackAccessService(db), compat, compat, compat, webhook))
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
