package aws

import (
	"github.com/alphauslabs/bluectl/cmds/cost/aws/adjustments"
	"github.com/alphauslabs/bluectl/cmds/cost/aws/attributes"
	"github.com/alphauslabs/bluectl/cmds/cost/aws/calculations"
	"github.com/alphauslabs/bluectl/cmds/cost/aws/calculator"
	"github.com/alphauslabs/bluectl/cmds/cost/aws/usage"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "AWS-specific cost-related subcommands",
		Long:  `AWS-specific cost-related subcommands.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		usage.GetCmd(), // compatibility
		usage.Cmd(),
		adjustments.Cmd(),
		attributes.Cmd(),
		calculations.Cmd(),
		calculator.Cmd(),
	)

	return cmd
}
