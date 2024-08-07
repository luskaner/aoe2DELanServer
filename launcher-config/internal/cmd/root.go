package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var rootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "config execute config-only tasks",
	Long:  "config execute config-only tasks as required by launcher directly or indirectly via the agent",
}

var Version string

func Execute() error {
	initSetUp()
	initRevert()
	rootCmd.PersistentFlags().StringVarP(&Version, "version", "v", Version, "Version")
	return rootCmd.Execute()
}
