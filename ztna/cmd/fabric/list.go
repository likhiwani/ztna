/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package fabric

import (
	logtrace "ztna-core/ztna/logtrace"
	"fmt"
	"strings"
	"time"

	fabric_rest_client "ztna-core/ztna/controller/rest_client"
	"ztna-core/ztna/controller/rest_client/circuit"
	"ztna-core/ztna/controller/rest_client/link"
	"ztna-core/ztna/controller/rest_client/router"
	"ztna-core/ztna/controller/rest_client/service"
	"ztna-core/ztna/controller/rest_client/terminator"
	"ztna-core/ztna/controller/rest_model"
	"ztna-core/ztna/ztna/cmd/api"
	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"github.com/openziti/foundation/v2/stringz"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

// newListCmd creates a command object for the "controller list" command
func newListCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "Lists various entities managed by the Ziti Controller",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			cmdhelper.CheckErr(err)
		},
	}

	newOptions := func() *api.Options {
		return &api.Options{CommonOptions: p()}
	}

	listCmd.AddCommand(newListCmdForEntityType("circuits", runListCircuits, newOptions()))
	listCmd.AddCommand(newListCmdForEntityType("links", runListLinks, newOptions()))
	listCmd.AddCommand(newListCmdForEntityType("routers", runListRouters, newOptions()))
	listCmd.AddCommand(newListCmdForEntityType("services", runListServices, newOptions()))
	listCmd.AddCommand(newListCmdForEntityType("terminators", runListTerminators, newOptions()))

	return listCmd
}

type listCommandRunner func(*api.Options) error

// newListCmdForEntityType creates the list command for the given entity type
func newListCmdForEntityType(entityType string, command listCommandRunner, options *api.Options, aliases ...string) *cobra.Command {
	logtrace.LogWithFunctionName()
	cmd := &cobra.Command{
		Use:     entityType + " <filter>?",
		Short:   "lists " + entityType + " managed by the Ziti Controller",
		Args:    cobra.MaximumNArgs(1),
		Aliases: aliases,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := command(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	cmd.Flags().BoolVar(&options.OutputCSV, "csv", false, "Output CSV instead of a formatted table")
	options.AddCommonFlags(cmd)

	return cmd
}

func runListCircuits(o *api.Options) error {
	logtrace.LogWithFunctionName()
	return WithFabricClient(o, func(client *fabric_rest_client.ZitiFabric) error {
		ctx, cancelF := o.GetContext()
		defer cancelF()
		result, err := client.Circuit.ListCircuits(&circuit.ListCircuitsParams{
			Filter:  o.GetFilter(),
			Context: ctx,
		})
		return outputResult(result, err, o, outputCircuits)
	})
}

func outputCircuits(o *api.Options, results *circuit.ListCircuitsOK) error {
	logtrace.LogWithFunctionName()
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"ID", "Client", "Service", "Terminator", "CreatedAt", "Path"})

	for _, entity := range results.Payload.Data {
		pathLabel := strings.Builder{}

		if path := entity.Path; path != nil {
			if len(path.Nodes) > 0 {
				pathLabel.WriteString("r/")
				pathLabel.WriteString(path.Nodes[0].Name)
			}
			for idx, node := range path.Nodes[1:] {
				linkEntity := path.Links[idx]
				pathLabel.WriteString(" -> l/")
				pathLabel.WriteString(linkEntity.ID)
				pathLabel.WriteString(" -> r/")
				pathLabel.WriteString(node.Name)
			}
		}

		t.AppendRow(table.Row{
			valOrDefault(entity.ID),
			entity.ClientID,
			entity.Service.Name,
			entity.Terminator.ID,
			time.Time(*entity.CreatedAt).UTC().Format(time.DateTime),
			pathLabel.String(),
		})
	}

	api.RenderTable(o, t, getPaging(results.Payload.Meta))

	return nil
}

func runListLinks(o *api.Options) error {
	logtrace.LogWithFunctionName()
	return WithFabricClient(o, func(client *fabric_rest_client.ZitiFabric) error {
		ctx, cancelF := o.GetContext()
		defer cancelF()
		result, err := client.Link.ListLinks(&link.ListLinksParams{
			Filter:  o.GetFilter(),
			Context: ctx,
		})
		return outputResult(result, err, o, outputLinks)
	})
}

func outputLinks(o *api.Options, results *link.ListLinksOK) error {
	logtrace.LogWithFunctionName()
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	columnConfigs := []table.ColumnConfig{
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
		{Number: 8, Align: text.AlignRight},
	}
	t.SetColumnConfigs(columnConfigs)
	t.AppendHeader(table.Row{"ID", "Dialer", "Acceptor", "Static Cost", "Src Latency", "Dst Latency", "State", "Status", "Full Cost"})

	for _, entity := range results.Payload.Data {
		id := valOrDefault(entity.ID)
		srcRouter := entity.SourceRouter.Name
		dstRouter := entity.DestRouter.Name
		staticCost := valOrDefault(entity.StaticCost)
		srcLatency := float64(valOrDefault(entity.SourceLatency)) / 1_000_000
		dstLatency := float64(valOrDefault(entity.DestLatency)) / 1_000_000
		state := valOrDefault(entity.State)
		down := valOrDefault(entity.Down)
		cost := valOrDefault(entity.Cost)

		status := "up"
		if down {
			status = "down"
		}

		t.AppendRow(table.Row{id, srcRouter, dstRouter, staticCost,
			fmt.Sprintf("%.1fms", srcLatency),
			fmt.Sprintf("%.1fms", dstLatency),
			state, status, cost})
	}

	api.RenderTable(o, t, getPaging(results.Payload.Meta))

	return nil
}

func runListTerminators(o *api.Options) error {
	logtrace.LogWithFunctionName()
	return WithFabricClient(o, func(client *fabric_rest_client.ZitiFabric) error {
		ctx, cancelF := o.GetContext()
		defer cancelF()

		result, err := client.Terminator.ListTerminators(&terminator.ListTerminatorsParams{
			Filter:  o.GetFilter(),
			Context: ctx,
		})
		return outputResult(result, err, o, outputTerminators)
	})
}

func outputTerminators(o *api.Options, result *terminator.ListTerminatorsOK) error {
	logtrace.LogWithFunctionName()
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"ID", "Service", "Router", "Binding", "Address", "Instance", "Cost", "Precedence", "Dynamic Cost", "Host ID"})

	for _, entity := range result.Payload.Data {
		id := valOrDefault(entity.ID)
		serviceName := entity.Service.Name
		routerName := entity.Router.Name
		binding := valOrDefault(entity.Binding)
		address := valOrDefault(entity.Address)
		instanceId := valOrDefault(entity.InstanceID)
		staticCost := valOrDefault(entity.Cost)
		precedence := valOrDefault(entity.Precedence)
		dynamicCost := valOrDefault(entity.DynamicCost)
		hostId := valOrDefault(entity.HostID)

		t.AppendRow(table.Row{id, serviceName, routerName, binding, address, instanceId, staticCost, precedence, dynamicCost, hostId})
	}

	api.RenderTable(o, t, getPaging(result.Payload.Meta))
	return nil
}

func runListServices(o *api.Options) error {
	logtrace.LogWithFunctionName()
	return WithFabricClient(o, func(client *fabric_rest_client.ZitiFabric) error {
		ctx, cancelF := o.GetContext()
		defer cancelF()

		result, err := client.Service.ListServices(&service.ListServicesParams{
			Filter:  o.GetFilter(),
			Context: ctx,
		})
		return outputResult(result, err, o, outputServices)
	})
}

func outputServices(o *api.Options, result *service.ListServicesOK) error {
	logtrace.LogWithFunctionName()
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"ID", "Name", "Terminator Strategy"})

	for _, entity := range result.Payload.Data {
		t.AppendRow(table.Row{
			valOrDefault(entity.ID),
			valOrDefault(entity.Name),
			valOrDefault(entity.TerminatorStrategy),
		})
	}

	api.RenderTable(o, t, getPaging(result.Payload.Meta))

	return nil
}

func runListRouters(o *api.Options) error {
	logtrace.LogWithFunctionName()
	return WithFabricClient(o, func(client *fabric_rest_client.ZitiFabric) error {
		ctx, cancelF := o.GetContext()
		defer cancelF()

		result, err := client.Router.ListRouters(&router.ListRoutersParams{
			Filter:  o.GetFilter(),
			Context: ctx,
		})
		return outputResult(result, err, o, outputRouters)
	})
}

func outputRouters(o *api.Options, result *router.ListRoutersOK) error {
	logtrace.LogWithFunctionName()
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"ID", "Name", "Online", "Cost", "No Traversal", "Disabled", "Version", "Listeners"})

	for _, entity := range result.Payload.Data {
		var version string
		if versionInfo := entity.VersionInfo; versionInfo != nil {
			version = fmt.Sprintf("%v on %v/%v", versionInfo.Version, versionInfo.Os, versionInfo.Arch)
		}
		var listeners []string
		for idx, listenerAddr := range entity.ListenerAddresses {
			addr := stringz.OrEmpty(listenerAddr.Address)
			listeners = append(listeners, fmt.Sprintf("%v: %v", idx+1, addr))
		}
		t.AppendRow(table.Row{
			valOrDefault(entity.ID),
			valOrDefault(entity.Name),
			valOrDefault(entity.Connected),
			valOrDefault(entity.Cost),
			valOrDefault(entity.NoTraversal),
			valOrDefault(entity.Disabled),
			version,
			strings.Join(listeners, "\n")})
	}

	api.RenderTable(o, t, getPaging(result.Payload.Meta))

	return nil
}

func getPaging(meta *rest_model.Meta) *api.Paging {
	logtrace.LogWithFunctionName()
	return &api.Paging{
		Limit:  *meta.Pagination.Limit,
		Offset: *meta.Pagination.Offset,
		Count:  *meta.Pagination.TotalCount,
	}
}

func outputResult[T any](val T, err error, o *api.Options, f func(o *api.Options, val T) error) error {
	logtrace.LogWithFunctionName()
	if err != nil {
		return err
	}
	if o.OutputJSONResponse {
		return nil
	}
	return f(o, val)
}

func valOrDefault[V any, T *V](val T) V {
	logtrace.LogWithFunctionName()
	var result V
	if val != nil {
		result = *val
	}
	return result
}
