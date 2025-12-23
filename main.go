package main

import (
	"go-reasonable-api/cmd/api"
	"go-reasonable-api/cmd/migrate"
	"go-reasonable-api/cmd/version"
	"go-reasonable-api/cmd/worker"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "go-reasonable-api",
		Short: "ZapAgenda - Backend API",
	}

	rootCmd.AddCommand(api.NewCommand())
	rootCmd.AddCommand(migrate.NewCommand())
	rootCmd.AddCommand(worker.NewCommand())
	rootCmd.AddCommand(version.NewCommand())

	cobra.CheckErr(rootCmd.Execute())
}
