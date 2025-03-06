package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"packwiz-web/internal/database"
	"strings"
)

var (
	modeOptions    = []string{"debug"}
	fmtModeOptions = strings.Join(modeOptions, " | ")

	bootstrap = &cobra.Command{
		Use:   fmt.Sprintf("bootstrap [%s]", fmtModeOptions),
		Short: "Initialize development data",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("invalid mode, must be one of: [%s]", fmtModeOptions)
			}

			for _, m := range modeOptions {
				if m == args[0] {
					return nil
				}
			}

			return fmt.Errorf("invalid mode, must be one of: [%s]", fmtModeOptions)

		},
		Run: func(cmd *cobra.Command, args []string) {
			choice := args[0]

			database.InitDb()

			switch choice {
			case "debug":
				database.SeedDebugData()
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(bootstrap)
}
