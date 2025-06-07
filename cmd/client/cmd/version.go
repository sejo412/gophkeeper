package cmd

import (
	"github.com/sejo412/gophkeeper/pkg/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the client version information",
	Long:  "\nPrint the client version information",
	Run: func(cmd *cobra.Command, args []string) {
		version := version.NewVersion()
		version.Print()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
