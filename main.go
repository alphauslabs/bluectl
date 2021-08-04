package main

import (
	"log"
	"os"

	_ "github.com/alphauslabs/blue-sdk-go/api"
	"github.com/alphauslabs/bluectl/cmds"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "bluectl",
		Short: "Command line interface for Alphaus services",
		Long: `Command line interface for Alphaus services.

The general form is bluectl <resource[ subresource...]> <action> [flags]. Most commands support the --raw-input
flag to be always in sync with the current feature set of the API in case the built-in flags don't support all
the possible input combinations. For beta APIs, we recommend you to use the --raw-input flag. See
https://alphauslabs.github.io/blueapidocs/ for the latest API reference.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if params.CleanOut {
				logger.SetPrefix(logger.PrefixNone)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}
)

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.PersistentFlags().StringVar(&params.AuthUrl, "auth-url", os.Getenv("ALPHAUS_AUTH_URL"), "authentication URL, defaults to $ALPHAUS_AUTH_URL if set")
	rootCmd.PersistentFlags().StringVar(&params.ClientId, "client-id", os.Getenv("ALPHAUS_CLIENT_ID"), "your client id, defaults to $ALPHAUS_CLIENT_ID")
	rootCmd.PersistentFlags().StringVar(&params.ClientSecret, "client-secret", os.Getenv("ALPHAUS_CLIENT_SECRET"), "your client secret, defaults to $ALPHAUS_CLIENT_SECRET")
	rootCmd.PersistentFlags().StringVar(&params.OutFile, "out", params.OutFile, "output file, if the command supports writing to file")
	rootCmd.PersistentFlags().StringVar(&params.OutFmt, "outfmt", "csv", "output format: json, csv, valid if --out is set")
	rootCmd.PersistentFlags().BoolVar(&params.CleanOut, "bare", params.CleanOut, "if true, set console output to barebones, easier for scripting")
	rootCmd.AddCommand(
		cmds.AccessTokenCmd(),
		cmds.WhoAmICmd(),
		cmds.AwsCostCmd(),
		cmds.AwsPayerCmd(),
		cmds.OpsCmd(),
	)
}

func main() {
	cobra.EnableCommandSorting = false
	log.SetOutput(os.Stdout)
	rootCmd.Execute()
}
