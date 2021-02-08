package main

import (
	"log"
	"os"

	"github.com/alphauslabs/bluectl/cmds"
	"github.com/alphauslabs/bluectl/logger"
	"github.com/alphauslabs/bluectl/params"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "bluectl",
		Short: "Command line interface for Alphaus Blue platform",
		Long:  `Command line interface for Alphaus Blue platform.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Error("invalid cmd, please run -h")
		},
	}
)

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.PersistentFlags().StringVar(&params.Target, "target", "ripple", "target for login: ripple, wave")
	rootCmd.PersistentFlags().StringVar(&params.ClientId, "client-id", os.Getenv("ALPHAUS_CLIENT_ID"), "your Alphaus client id")
	rootCmd.PersistentFlags().StringVar(&params.ClientSecret, "client-secret", os.Getenv("ALPHAUS_CLIENT_SECRET"), "your Alphaus client secret")
	rootCmd.PersistentFlags().StringVar(&params.OutFile, "out", params.OutFile, "output file, if the command supports writing to file")
	rootCmd.PersistentFlags().StringVar(&params.OutFmt, "out-fmt", params.OutFmt, "output format: json, csv, valid if --out is set")
	rootCmd.AddCommand(
		cmds.AccessTokenCmd(),
	)
}

func main() {
	log.SetOutput(os.Stdout)
	rootCmd.Execute()
}
