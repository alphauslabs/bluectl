package cmds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/alphauslabs/blue-sdk-go/org/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func OrgRegisterCmd() *cobra.Command {
	var (
		passwd string
		desc   string
	)

	cmd := &cobra.Command{
		Use:   "create <email>",
		Short: "Create a new organization",
		Long:  `Create a new organization.`,
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
				fnerr(fmt.Errorf("<email> cannot be empty"))
				return
			}

			if passwd == "" && !cmd.Flag("password").Changed {
				fmt.Println("Password will be generated if empty.")
				fmt.Print("Password: ")
				pw1, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					fnerr(err)
					return
				}

				fmt.Print("\nConfirm password: ")
				pw2, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					fnerr(err)
					return
				}

				switch {
				case string(pw1) == string(pw2):
					passwd = string(pw1)
					fmt.Println("")
				default:
					fmt.Println("\nInvalid password.")
					return
				}
			}

			if desc == "" {
				fmt.Print("Description: ")
				fmt.Scanln(&desc)
				if desc == "" {
					fnerr(fmt.Errorf("Description is empty."))
					return
				}
			}

			hc := &http.Client{Timeout: 60 * time.Second}
			u := "https://api.alphaus.cloud/m/blue/org/v1"
			entry := make(map[string]interface{})
			entry["email"] = args[0]
			entry["password"] = passwd
			entry["description"] = desc
			payload, _ := json.Marshal(entry)

			logger.Info(string(payload))
			return

			r, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(payload))
			if err != nil {
				fnerr(err)
				return
			}

			resp, err := hc.Do(r)
			if err != nil {
				fnerr(err)
				return
			}

			if (resp.StatusCode / 100) != 2 {
				fnerr(fmt.Errorf(resp.Status))
				return
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fnerr(err)
				return
			}

			logger.Info(string(body))
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&passwd, "password", passwd, "your org password, generated if set to empty")
	cmd.Flags().StringVar(&desc, "description", desc, "your org description (e.g. org name)")
	return cmd
}

func OrgGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Print information about your organization",
		Long:  `Print information about your organization.`,
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
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.OrgService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := org.NewClient(ctx, &org.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			resp, err := client.GetOrg(ctx, &org.GetOrgRequest{})
			if err != nil {
				fnerr(err)
				return
			}

			switch {
			case params.OutFmt == "json":
				b, _ := json.Marshal(resp)
				logger.Info(string(b))
			default:
				logger.Info(resp)
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func OrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Subcommand for Organization operations",
		Long:  `Subcommand for Organization operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		OrgRegisterCmd(),
		OrgGetCmd(),
	)

	return cmd
}
