package main

import (
	"log"
	"os"

	"github.com/alphauslabs/bluectl/cmds"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "bluectl",
		Short: "Command line interface for Alphaus services",
		Long:  `Command line interface for Alphaus services.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Error("invalid cmd, please run -h")
		},
	}
)

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.PersistentFlags().StringVar(&params.AuthUrl, "auth-url", os.Getenv("ALPHAUS_AUTH_URL"), "Alphaus authentication URL, defaults to $ALPHAUS_AUTH_URL")
	rootCmd.PersistentFlags().StringVar(&params.ClientId, "client-id", os.Getenv("ALPHAUS_CLIENT_ID"), "your client id, defaults to $ALPHAUS_CLIENT_ID")
	rootCmd.PersistentFlags().StringVar(&params.ClientSecret, "client-secret", os.Getenv("ALPHAUS_CLIENT_SECRET"), "your client secret, defaults to $ALPHAUS_CLIENT_SECRET")
	rootCmd.PersistentFlags().StringVar(&params.OutFile, "out", params.OutFile, "output file, if the command supports writing to file")
	rootCmd.PersistentFlags().StringVar(&params.OutFmt, "outfmt", "csv", "output format: json, csv, valid if --out is set")
	rootCmd.AddCommand(
		cmds.WhoAmICmd(),
		cmds.AccessTokenCmd(),
	)
}

func main() {
	log.SetOutput(os.Stdout)
	rootCmd.Execute()
}
