package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"packwiz-web/internal/config"
)

var (
	rootCmd = &cobra.Command{
		Use: config.C.Name,
		Long: fmt.Sprintf("%s %s - Packwiz Web Service",
			config.C.Name,
			config.C.Version),
	}
)

func Execute() {
	_ = rootCmd.Execute()
}
