package cmd

import (
	"net"
	"strconv"

	"github.com/sejo412/gophkeeper/internal/client"
	"github.com/sejo412/gophkeeper/internal/constants"
	"github.com/spf13/cobra"
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register user",
	Long: `
          Register new user
          !!!!!!!!!!!!!!!!!
          !!! Attention !!!
          !!!!!!!!!!!!!!!!!
All certificates will be overwritten!
`,
	Run: func(cmd *cobra.Command, args []string) {
		c := client.NewClient(
			client.Config{
				PublicAddress: publicHost,
				CacheDir:      cacheDir,
			},
		)
		if err := c.Register(userName); err != nil {
			panic(err)
		}
	},
}

func init() {
	defaultPublicServer := net.JoinHostPort(constants.DefaultServerHost, strconv.Itoa(constants.DefaultPublicPort))
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().StringVarP(
		&publicHost, "server", "s", defaultPublicServer, "public server address",
	)
	registerCmd.Flags().StringVarP(&cacheDir, "dir", "d", client.DefaultCacheDir(), "cache directory")
	registerCmd.Flags().StringVarP(&userName, "user", "u", "", "user name")
	_ = registerCmd.MarkFlagRequired("user")
}
