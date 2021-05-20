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
	rootCmd.PersistentFlags().StringVar(&params.RippleClientId, "ripple-client-id", os.Getenv("ALPHAUS_RIPPLE_CLIENT_ID"), "your Ripple client id")
	rootCmd.PersistentFlags().StringVar(&params.RippleClientSecret, "ripple-client-secret", os.Getenv("ALPHAUS_RIPPLE_CLIENT_SECRET"), "your Ripple client secret")
	rootCmd.PersistentFlags().StringVar(&params.WaveClientId, "wave-client-id", os.Getenv("ALPHAUS_WAVE_CLIENT_ID"), "your Wave client id")
	rootCmd.PersistentFlags().StringVar(&params.WaveClientSecret, "wave-client-secret", os.Getenv("ALPHAUS_WAVE_CLIENT_SECRET"), "your Wave client secret")
	rootCmd.PersistentFlags().StringVar(&params.OutFile, "out", params.OutFile, "output file, if the command supports writing to file")
	rootCmd.PersistentFlags().StringVar(&params.OutFmt, "outfmt", "csv", "output format: json, csv, valid if --out is set")
	rootCmd.AddCommand(
		cmds.RippleCmd(),
		cmds.WaveCmd(),
	)
}

func main() {
	log.SetOutput(os.Stdout)
	rootCmd.Execute()
}
