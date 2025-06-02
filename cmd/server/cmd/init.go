package cmd

import (
	"github.com/sejo412/gophkeeper/internal/server"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new server",
	Long: `
         Initialize a new server
            !!!!!!!!!!!!!!!!!
            !!! Attention !!!
            !!!!!!!!!!!!!!!!!
All data and certificates will be overwritten!
`,
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewServer(server.Config{
			CacheDir: cacheDir,
		})
		if err := s.Init(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
