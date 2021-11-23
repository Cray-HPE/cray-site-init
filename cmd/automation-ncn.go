package cmd

import (
	"github.com/Cray-HPE/cray-site-init/pkg/kubernetes"
	"github.com/spf13/cobra"
)

// Common vars.
var (
	kubernetesClient *kubernetes.UtilsClient

	action string
	ncn    string
)

var automateNCNCommand = &cobra.Command{
	Use:   "ncn",
	Short: "tools used to automate NCN activities",
	Long:  "A series of subcommands that automates NCN administrative tasks.",
}

func init() {
	automateCommand.AddCommand(automateNCNCommand)
	automateCommand.DisableAutoGenTag = true
}
