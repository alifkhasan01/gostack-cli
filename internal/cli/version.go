package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// Commit and BuildDate are optionally set at build time.
var (
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the GoStack CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GoStack CLI %s\n", Version)
		fmt.Printf("  Commit    : %s\n", Commit)
		fmt.Printf("  Built     : %s\n", BuildDate)
		fmt.Printf("  Go        : %s\n", runtime.Version())
		fmt.Printf("  Platform  : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
