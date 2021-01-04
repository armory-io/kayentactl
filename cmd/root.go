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
	"fmt"
	"os"

	"github.com/armory-io/kayentactl/internal/logger"
	"github.com/armory-io/kayentactl/pkg/kayenta"
	"github.com/fatih/color"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	verbosity string
	noColor   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kayentactl",
	Short: "",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initLogs(verbosity); err != nil {
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

	fmt.Printf("%v\n", color.HiMagentaString(kayenta.AsciiKayenta))
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		log.Fatal("Could not parse CLI arguments. Exiting.")
	}
}

var kayentaURL string

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&kayentaURL, "kayenta-url", "u", "http://localhost:8090", "kayenta url")
	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", log.InfoLevel.String(), "log level (debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable output colors")

}
