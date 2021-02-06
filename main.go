package main

import (
	"log"
	"os"

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
	// rootCmd.AddCommand(
	// 	cmds.LsDbCmd(),
	// 	cmds.LsQueuedCmd(),
	// 	cmds.CheckEmptyCmd(),
	// 	cmds.CleanDBCmd(),
	// 	cmds.CleanCompanyCmd(),
	// 	cmds.UserInfoCmd(),
	// 	cmds.CompanyInfoCmd(),
	// 	cmds.TotalsCmd(),
	// 	cmds.FinishInvoiceCmd(),
	// 	cmds.CopyReportsCmd(),
	// 	cmds.DynamoTableSyncCmd(),
	// 	cmds.TokenCmd(),
	// 	cmds.CheckPayerS3Cmd(),
	// 	cmds.SpannerToBqCmd(),
	// 	cmds.QueryDailyCmd(),
	// 	cmds.ValidateDetailsCmd(),
	// 	cmds.ValidateFeesCmd(),
	// 	cmds.SnapshotsCmd(),
	// 	cmds.GenSqlCmd(),
	// 	cmds.SetPayerMspCmd(),
	// 	cmds.ReInvoiceCmd(),
	// 	cmds.ResetUserMFA(),
	// 	cmds.OthersCmd(),
	// 	cmds.TestCmd(),
	// )
}

func main() {
	log.SetOutput(os.Stdout)
	rootCmd.Execute()
}
