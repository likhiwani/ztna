package fabric

import (
	logtrace "ztna-core/ztna/logtrace"
	"context"

	"ztna-core/ztna/controller/rest_client/raft"
	"ztna-core/ztna/controller/rest_model"
	"ztna-core/ztna/ztna/cmd/api"
	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"ztna-core/ztna/ztna/util"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// newRaftCmd creates a command object for the "controller raft" command
func newRaftCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	cmd := &cobra.Command{
		Use:   "raft",
		Short: "Raft operations",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			cmdhelper.CheckErr(err)
		},
	}

	cmd.AddCommand(newRaftListMembersCmd(p))
	cmd.AddCommand(newRaftAddMemberCmd(p))
	cmd.AddCommand(newRaftRemoveMemberCmd(p))
	cmd.AddCommand(newRaftTransferLeadershipCmd(p))

	return cmd
}

func newRaftListMembersCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &raftListMembersAction{
		Options: api.Options{CommonOptions: p()},
	}

	cmd := &cobra.Command{
		Use:   "list-members",
		Short: "list cluster members and their status",
		Args:  cobra.ExactArgs(0),
		RunE:  action.run,
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	action.AddCommonFlags(cmd)

	return cmd
}

type raftListMembersAction struct {
	api.Options
}

func (self *raftListMembersAction) run(cmd *cobra.Command, _ []string) error {
	logtrace.LogWithFunctionName()
	self.Cmd = cmd
	client, err := util.NewFabricManagementClient(self)
	if err != nil {
		return err
	}
	members, err := client.Raft.RaftListMembers(&raft.RaftListMembersParams{
		Context: context.Background(),
	})
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"Id", "Address", "Voter", "Leader", "Version", "Connected", "ReadOnly"})
	for _, m := range members.Payload.Data {
		t.AppendRow(table.Row{*m.ID, *m.Address, *m.Voter, *m.Leader, *m.Version, *m.Connected, m.ReadOnly != nil && *m.ReadOnly})
	}
	api.RenderTable(&api.Options{
		CommonOptions: self.CommonOptions,
	}, t, nil)
	return nil
}

func newRaftAddMemberCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &raftAddMemberAction{
		Options: api.Options{CommonOptions: p()},
	}

	cmd := &cobra.Command{
		Use:   "add-member <address>",
		Short: "add cluster member",
		Args:  cobra.ExactArgs(1),
		RunE:  action.run,
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	action.AddCommonFlags(cmd)
	cmd.Flags().BoolVar(&action.nonVoting, "non-voting", false, "Allows adding a non-voting member to the cluster")

	return cmd
}

type raftAddMemberAction struct {
	api.Options
	nonVoting bool
}

func (self *raftAddMemberAction) run(cmd *cobra.Command, args []string) error {
	logtrace.LogWithFunctionName()
	self.Cmd = cmd
	client, err := util.NewFabricManagementClient(self)
	if err != nil {
		return err
	}

	isVoter := !self.nonVoting

	_, err = client.Raft.RaftMemberAdd(&raft.RaftMemberAddParams{
		Context: context.Background(),
		Member: &rest_model.RaftMemberAdd{
			Address: &args[0],
			IsVoter: &isVoter,
		},
	})

	return err
}

func newRaftRemoveMemberCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &raftRemoveMemberAction{
		Options: api.Options{CommonOptions: p()},
	}

	cmd := &cobra.Command{
		Use:   "remove-member <cluster member id>",
		Short: "remove cluster member",
		Args:  cobra.ExactArgs(1),
		RunE:  action.run,
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	action.AddCommonFlags(cmd)

	return cmd
}

type raftRemoveMemberAction struct {
	api.Options
}

func (self *raftRemoveMemberAction) run(cmd *cobra.Command, args []string) error {
	logtrace.LogWithFunctionName()
	self.Cmd = cmd

	client, err := util.NewFabricManagementClient(self)
	if err != nil {
		return err
	}

	_, err = client.Raft.RaftMemberRemove(&raft.RaftMemberRemoveParams{
		Context: context.Background(),
		Member: &rest_model.RaftMemberRemove{
			ID: &args[0],
		},
	})

	return err
}

func newRaftTransferLeadershipCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &raftTransferLeadershipAction{
		Options: api.Options{CommonOptions: p()},
	}

	cmd := &cobra.Command{
		Use:   "transfer-leadership [cluster member id]?",
		Short: "transfer cluster leadership to another member",
		Long:  "transfer cluster leadership to another member. If a node id is specified, leadership will be transferred to that node",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  action.run,
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	action.AddCommonFlags(cmd)

	return cmd
}

type raftTransferLeadershipAction struct {
	api.Options
}

func (self *raftTransferLeadershipAction) run(cmd *cobra.Command, args []string) error {
	logtrace.LogWithFunctionName()
	self.Cmd = cmd

	client, err := util.NewFabricManagementClient(self)
	if err != nil {
		return err
	}

	newLeader := ""

	if len(args) > 0 {
		newLeader = args[0]
	}

	_, err = client.Raft.RaftTransferLeadership(&raft.RaftTransferLeadershipParams{
		Context: context.Background(),
		Member: &rest_model.RaftTransferLeadership{
			NewLeaderID: newLeader,
		},
	})

	return err
}
