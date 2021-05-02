/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/armory-io/kayentactl/internal/options"

	"github.com/armory-io/kayentactl/cmd/accounts"

	"github.com/armory-io/kayentactl/cmd/analysis"
	"github.com/armory-io/kayentactl/cmd/version"

	"github.com/armory-io/kayentactl/internal/logger"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	verbosity  string
	noColor    bool
	kayentaURL string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kayentactl",
	Short: "",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		globals, _ := options.Globals(cmd)
		if err := initLogs(globals.Verbosity); err != nil {
			return err
		}
		return nil
	},
}

func initLogs(level string) error {
	log.SetOutput(os.Stdout)
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	var formatter log.Formatter = &logger.ColorizedLogger{}
	if !noColor {
		formatter = &logger.PlainLogger{}
	}
	log.SetFormatter(formatter)

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		log.Fatal("Could not parse CLI arguments. Exiting.")
	}
}

func init() {
	analysis.Configure(rootCmd)
	accounts.Configure(rootCmd)
	version.Configure(rootCmd)
	// global options are added by an external pacakge so that they can be
	// managed from a single source and used across all sub-commands. this
	// ensures that the logic for getting then stays consistent
	options.ConfigureGlobals(rootCmd)
}
