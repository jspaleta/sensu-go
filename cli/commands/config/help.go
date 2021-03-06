package config

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Modify sensuctl configuration",
	}

	// Add sub-commands
	cmd.AddCommand(
		SetEnvCommand(cli),
		SetFormatCommand(cli),
		SetOrgCommand(cli),
		ViewCommand(cli),
	)

	return cmd
}
