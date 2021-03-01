package options

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func ConfigureGlobals(cmd *cobra.Command) {
	persistentFlags := cmd.PersistentFlags()
	persistentFlags.StringP("kayenta-url", "u", "http://localhost:8090", "kayenta url")
	persistentFlags.StringP("verbosity", "v", log.InfoLevel.String(), "log level (debug, info, warn, error, fatal, panic)")
	persistentFlags.Bool("no-color", false, "disable output colors")
}

type GlobalOptions struct {
	KayentaURL, Verbosity string
	NoColor               bool
}

func Globals(cmd *cobra.Command) (*GlobalOptions, error) {
	flags := cmd.PersistentFlags()

	kayentaURL, _ := flags.GetString("kayenta-url")
	verbosity, _ := flags.GetString("verbosity")
	noColor, _ := flags.GetBool("no-color")

	return &GlobalOptions{
		KayentaURL: kayentaURL,
		Verbosity:  verbosity,
		NoColor:    noColor,
	}, nil
}
