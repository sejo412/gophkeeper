package cmd

import (
	"os"

	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/sejo412/gophkeeper/internal/server"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Start GophKeeper server application",
	Long:  "\nStart GophKeeper server application",
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewServer(server.Config{
			PublicPort:  publicPort,
			PrivatePort: privatePort,
			CacheDir:    cacheDir,
			DNSNames:    dnsNames,
		})
		if err := s.Start(); err != nil {
			panic(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&publicPort, "public-port", "p", constants.DefaultPublicPort,
		"Public port to listen on")
	rootCmd.Flags().IntVarP(&privatePort, "private-port", "s", constants.DefaultPrivatePort,
		"Private port to listen on (with TLS)")
	rootCmd.Flags().StringSliceVar(&dnsNames, "dns", constants.DefaultDNSNames, "DNS names to serve")
	rootCmd.PersistentFlags().StringVarP(&cacheDir, "dir", "d", server.DefaultCacheDir(),
		"Cache directory to save certificates and database")
}
