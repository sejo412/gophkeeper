package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/sejo412/gophkeeper/internal/client"
	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "GophKeeper client application",
	Long:  "\nGophKeeper client application",
	Run: func(cmd *cobra.Command, args []string) {
		c := client.NewClient(
			client.Config{
				PrivateAddress: privateHost,
				CacheDir:       cacheDir,
			},
		)
		if err := c.Run(); err != nil {
			fmt.Println(err)
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
	defaultPrivateHost := net.JoinHostPort(constants.DefaultServerHost, strconv.Itoa(constants.DefaultPrivatePort))
	rootCmd.Flags().StringVarP(&privateHost, "server", "s", defaultPrivateHost, "private server address")
	rootCmd.PersistentFlags().StringVarP(&cacheDir, "dir", "d", client.DefaultCacheDir(), "cache directory")
}
