package edge

import (
	"io"

	"ztna-core/ztna/logtrace"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"

	"github.com/spf13/cobra"
)

func newDbCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management operations for the Ziti Edge Controller",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			cmdhelper.CheckErr(err)
		},
	}

	cmd.AddCommand(newDbSnapshotCmd(out, errOut))
	cmd.AddCommand(newDbCheckIntegrityCmd(out, errOut))
	cmd.AddCommand(newDbCheckIntegrityStatusCmd(out, errOut))

	return cmd
}
