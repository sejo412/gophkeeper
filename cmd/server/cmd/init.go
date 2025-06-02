package cmd

import (
	"github.com/sejo412/gophkeeper/internal/config"
	"github.com/sejo412/gophkeeper/internal/server"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new server.",
	Long: `Initialize a new server.

!!! Attention !!! All data and certificates will be overwritten!
`,
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewServer(config.ServerConfig{
			PublicPort:  publicPort,
			PrivatePort: privatePort,
			CacheDir:    cacheDir,
		})
		if err := s.Init(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
