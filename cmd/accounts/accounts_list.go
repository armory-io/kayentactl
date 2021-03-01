package accounts

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/armory-io/kayentactl/internal/options"

	"github.com/armory-io/kayentactl/pkg/kayenta"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	types, outFormat string
)

var shorthandMap = map[string]string{
	"metrics": "METRICS_STORE",
	"config":  "CONFIGURATION_STORE",
	"object":  "OBJECT_STORE",
}

// getCmd represents the get command
var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		globals, _ := options.Globals(cmd)
		kc := kayenta.NewDefaultClient(kayenta.ClientBaseURL(globals.KayentaURL))
		credentials, err := kc.GetCredentials()
		if err != nil {
			log.Fatalf("could not get list of available accounts: %s", err.Error())
		}

		// TODO: add type filtering

		var (
			output []byte
		)
		switch outFormat {
		case "json":
			output, err = outputJson(credentials)
		case "pretty":
			// TODO: implement a pretty output instead of just falling back to json
			output, err = outputPretty(credentials)
		}

		if err != nil {
			log.Fatalf(err.Error())
		}

		fmt.Fprint(os.Stdout, string(output))
	},
}

func outputJson(creds []kayenta.AccountCredential) ([]byte, error) {
	return json.MarshalIndent(creds, "", "  ")
}

func outputPretty(creds []kayenta.AccountCredential) ([]byte, error) {
	return outputJson(creds)
}

func init() {
	accountsCmd.AddCommand(accountListCmd)
	accountListCmd.Flags().StringVarP(&types, "types", "t", "", "filter by account type")
	accountListCmd.Flags().StringVarP(&outFormat, "output", "o", "pretty", "output format: json|pretty")

}
