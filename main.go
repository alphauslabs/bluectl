package main

import (
	"log"
	"os"

	"github.com/alphauslabs/bluectl/cmds"
	"github.com/mobingi/gosdk/pkg/util/simplelog"
	"github.com/spf13/cobra"
)

var (
	outfmt string

	rootCmd = &cobra.Command{
		Use:   "bluectl",
		Short: "Command line interface for Alphaus Blue platform",
		Long:  `Command line interface for Alphaus Blue platform.`,
		Run: func(cmd *cobra.Command, args []string) {
			simplelog.Error("invalid cmd, please run -h")
		},
	}
)

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.PersistentFlags().StringVar(&outfmt, "output-fmt", outfmt, "output format: json, csv")
	rootCmd.AddCommand(
		cmds.AccessTokenCmd(),
	)
}

func main() {
	log.SetOutput(os.Stdout)
	rootCmd.Execute()
}
