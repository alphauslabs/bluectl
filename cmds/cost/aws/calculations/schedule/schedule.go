package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List calculation schedules",
		Long:  `List calculation schedules.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			ctx := context.Background()
			con, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: con})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			resp, err := client.ListCalculationsSchedules(ctx, &cost.ListCalculationsSchedulesRequest{Vendor: "aws"})
			if err != nil {
				fnerr(err)
				return
			}

			switch {
			case params.OutFmt == "json":
				b, _ := json.Marshal(resp)
				if err != nil {
					fnerr(err)
					return
				}

				fmt.Printf("%v", string(b))
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.SetAutoFormatHeaders(false)
				table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
				table.SetColWidth(100)
				table.SetBorder(false)
				table.SetHeaderLine(false)
				table.SetColumnSeparator("")
				table.SetTablePadding("  ")
				table.SetNoWhiteSpace(true)
				table.SetHeader([]string{
					"ID",
					"SCHEDULE",
					"SCHEDULE_MACRO",
					"TARGET_MONTH",
					"NEXT_RUN",
					"NOTIFICATION_CHANNEL",
					"DRYRUN",
				})

				if len(resp.Schedules) > 0 {
					table.Append([]string{
						resp.Schedules[0].Id,
						resp.Schedules[0].Schedule,
						resp.Schedules[0].ScheduleMacro,
						resp.Schedules[0].TargetMonth,
						resp.Schedules[0].NextRun,
						resp.Schedules[0].NotificationChannel,
						fmt.Sprintf("%v", resp.Schedules[0].DryRun),
					})
				}

				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func CreateCmd() *cobra.Command {
	var (
		rawInput   string
		notifyChan string
		dryrun     bool
	)

	cmd := &cobra.Command{
		Use:   "create <cron>",
		Short: "Create a calculation schedule",
		Long: `Create a calculation schedule in cron format. At the moment, only one schedule is permitted
per Ripple account. Enclose your input with either '' or "".

For example, if you want to schedule your calculation every 3rd day of the month:
  bluectl cost aws calculation schedule create "0 0 3 * *"

You can get the notification channel id by using the command:
  bluectl notification channels list`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			if len(args) == 0 && rawInput == "" {
				fnerr(fmt.Errorf("id cannot be empty"))
				return
			}

			ctx := context.Background()
			con, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: con})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			var r cost.CreateCalculationsScheduleRequest
			switch {
			case rawInput != "":
				err := json.Unmarshal([]byte(rawInput), &r)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				r = cost.CreateCalculationsScheduleRequest{
					Vendor:   "aws",
					Schedule: args[0],
					Force:    true, // default at the moment
					DryRun:   dryrun,
				}

				if notifyChan != "" {
					r.NotificationChannel = notifyChan
				}
			}

			resp, err := client.CreateCalculationsSchedule(ctx, &r)
			if err != nil {
				fnerr(err)
				return
			}

			b, _ := json.Marshal(resp)
			logger.Info(string(b))
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://labs.alphaus.cloud/blueapidocs/#/Cost/Cost_CreateCalculationsSchedule")
	cmd.Flags().StringVar(&notifyChan, "notification-channel", notifyChan, "notification channel id; if empty, creates a channel using your email")
	cmd.Flags().BoolVar(&dryrun, "dryrun", dryrun, "if true, simulate notification only, no actual calculation")
	return cmd
}

func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <id|-|.>",
		Short: "Delete calculation schedules",
		Long:  `Delete calculation schedules. Accepts an id, or '-', or '.', which means all.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			if len(args) == 0 {
				fnerr(fmt.Errorf("id cannot be empty"))
				return
			}

			id := args[0]
			if id == "." {
				id = "*"
			}

			ctx := context.Background()
			con, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: con})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			r := cost.DeleteCalculationsScheduleRequest{Vendor: "aws", Id: id}
			_, err = client.DeleteCalculationsSchedule(ctx, &r)
			if err != nil {
				fnerr(err)
				return
			}

			logger.Infof("[%v] deleted", args[0])
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Subcommand for calculation schedules",
		Long:  `Subcommand for calculation schedules.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		ListCmd(),
		CreateCmd(),
		DeleteCmd(),
	)

	return cmd
}
