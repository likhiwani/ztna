package database

import (
	"fmt"
	"io"

	"github.com/cosmic-cloak/ztna/ztna/util"
	"github.com/openziti/ziti-db-explorer/cmd/ziti-db-explorer/zdecli"
	"github.com/spf13/cobra"
)

func NewCmdDb(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := util.NewEmptyParentCmd("db", "Interact with Ziti database files")

	exploreCmd := &cobra.Command{
		Use:   "explore <ctrl.db>|help|version",
		Short: "Interactive CLI to explore Ziti database files",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := zdecli.Run("ziti db explore", args[0]); err != nil {
				_, _ = errOut.Write([]byte(fmt.Sprintf("Error: %s", err)))
			}
		},
	}

	cmd.AddCommand(exploreCmd)
	cmd.AddCommand(NewCompactAction())
	cmd.AddCommand(NewDiskUsageAction())
	cmd.AddCommand(NewAddDebugAdminAction())
	cmd.AddCommand(NewAnonymizeAction())

	return cmd
}
