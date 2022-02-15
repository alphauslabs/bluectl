package main

import (
	"log"
	"os"
	"path/filepath"

	_ "github.com/alphauslabs/blue-sdk-go/api"
	"github.com/alphauslabs/bluectl/cmds"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/fatih/color"
	tomlv2 "github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

var (
	bold = color.New(color.Bold).SprintFunc()

	rootCmd = &cobra.Command{
		Use:   "bluectl",
		Short: bold("bluectl") + " - Command line interface for Alphaus services",
		Long: bold("bluectl") + ` - Command line interface for Alphaus services.
Copyright (c) 2021-2022 Alphaus Cloud, Inc. All rights reserved.

The general form is ` + bold("bluectl <resource[ subresource...]> <action> [flags]") + `. Most commands support the ` + bold("--raw-input") + `
flag to be always in sync with the current feature set of the API in case the built-in flags don't support all
the possible input combinations. For beta APIs, we recommend you to use the ` + bold("--raw-input") + ` flag. See
https://alphauslabs.github.io/blueapidocs/ for the latest API reference.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if params.CleanOut {
				logger.SetPrefix(logger.PrefixNone)
			}

			home, _ := os.UserHomeDir()
			cfgfile := filepath.Join(home, ".config", "alphaus", "config.toml")
			_, err := os.Stat(cfgfile)
			if err == nil {
				if params.AuthProfile == "" {
					params.AuthProfile = "default"
				}
			}

			if params.AuthProfile != "" {
				b, err := os.ReadFile(cfgfile)
				if err != nil {
					logger.Error(err)
					os.Exit(1)
				}

				var cfg map[string]map[string]string
				err = tomlv2.Unmarshal(b, &cfg)
				if err != nil {
					logger.Error(err)
					os.Exit(1)
				}

				if v, ok := cfg[params.AuthProfile]; ok {
					if _, ok = v["client-id"]; ok {
						params.ClientId = v["client-id"]
					}

					if _, ok = v["client-secret"]; ok {
						params.ClientSecret = v["client-secret"]
					}

					if _, ok = v["auth-url"]; ok {
						params.AuthUrl = v["auth-url"]
					}
				} else {
					logger.Errorf("[%v] not found in %v", params.AuthProfile, cfgfile)
					os.Exit(1)
				}
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
	rootCmd.PersistentFlags().StringVar(&params.AuthProfile, "profile", params.AuthProfile, "profile name in ~/.config/alphaus/config.toml, default is [default]")
	rootCmd.PersistentFlags().StringVar(&params.AuthUrl, "auth-url", os.Getenv("ALPHAUS_AUTH_URL"), "authentication URL, defaults to $ALPHAUS_AUTH_URL if set")
	rootCmd.PersistentFlags().StringVar(&params.ClientId, "client-id", os.Getenv("ALPHAUS_CLIENT_ID"), "your client id, defaults to $ALPHAUS_CLIENT_ID")
	rootCmd.PersistentFlags().StringVar(&params.ClientSecret, "client-secret", os.Getenv("ALPHAUS_CLIENT_SECRET"), "your client secret, defaults to $ALPHAUS_CLIENT_SECRET")
	rootCmd.PersistentFlags().StringVar(&params.OutFile, "out", params.OutFile, "output file, if the command supports writing to file")
	rootCmd.PersistentFlags().StringVar(&params.OutFmt, "outfmt", "csv", "output format: json, csv, valid if --out is set")
	rootCmd.PersistentFlags().BoolVar(&params.CleanOut, "bare", params.CleanOut, "if true, set console output to barebones, easier for scripting")
	rootCmd.AddCommand(
		cmds.AccessTokenCmd(),
		cmds.WhoAmICmd(),
		cmds.OrgCmd(),
		cmds.IamCmd(),
		cmds.IdpCmd(),
		cmds.CrossAcctAccessCmd(),
		cmds.CostCmd(),
		cmds.AwsTagsCmd(),
		cmds.AwsPayerCmd(),
		cmds.AwsBillCmd(),
		cmds.NotificationCmd(),
		cmds.OpsCmd(),
		cmds.KvCmd(),
	)
}

func main() {
	cobra.EnableCommandSorting = false
	log.SetOutput(os.Stdout)
	rootCmd.Execute()
}
