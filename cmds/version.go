package cmds

import (
	"fmt"

	"github.com/alphauslabs/bluectl/params"
	"github.com/spf13/cobra"
)

func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get current version",
		Long:  `Get current version.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("bluectl", params.Version)
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
