package version

import (
	"fmt"

	"go-reasonable-api/support/version"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Full())
		},
	}
}
