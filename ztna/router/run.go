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

package router

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"ztna-core/ztna/common"
	"ztna-core/ztna/common/version"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router"
	"ztna-core/ztna/router/debugops"
	"ztna-core/ztna/router/xgress"
	"ztna-core/ztna/router/xgress_edge"
	"ztna-core/ztna/router/xgress_edge_transport"
	"ztna-core/ztna/router/xgress_edge_tunnel"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/agent"
	"github.com/openziti/foundation/v2/debugz"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewRunCmd() *cobra.Command {
	logtrace.LogWithFunctionName()
	var runCmd = &cobra.Command{
		Use:   "run <config>",
		Short: "Run router configuration",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}

	//flags are added to an internal map and read later on, see getFlags()
	runCmd.Flags().BoolP("extend", "e", false, "force the router on startup to extend enrollment certificates")
	return runCmd
}

func run(cmd *cobra.Command, args []string) {
	logtrace.LogWithFunctionName()
	startLogger := logrus.WithField("version", version.GetVersion()).
		WithField("go-version", version.GetGoVersion()).
		WithField("os", version.GetOS()).
		WithField("arch", version.GetArchitecture()).
		WithField("build-date", version.GetBuildDate()).
		WithField("revision", version.GetRevision()).
		WithField("configFile", args[0])

	config, err := router.LoadConfig(args[0])
	if err != nil {
		startLogger.WithError(err).Error("error loading ziti router config")
		panic(err)
	}
	config.SetFlags(getFlags(cmd))

	startLogger = startLogger.WithField("routerId", config.Id.Token)
	startLogger.Info("starting ziti router")

	r := router.Create(config, version.GetCmdBuildInfo())

	xgressEdgeFactory := xgress_edge.NewFactory(config, r, r.GetStateManager())
	xgress.GlobalRegistry().Register(common.EdgeBinding, xgressEdgeFactory)
	if err := r.RegisterXrctrl(xgressEdgeFactory); err != nil {
		logrus.WithError(err).Panic("error registering edge in framework")
	}

	xgressEdgeTransportFactory := xgress_edge_transport.NewFactory()
	xgress.GlobalRegistry().Register(xgress_edge_transport.BindingName, xgressEdgeTransportFactory)

	xgressEdgeTunnelFactory := xgress_edge_tunnel.NewFactory(r, config, r.GetStateManager())
	xgress.GlobalRegistry().Register(common.TunnelBinding, xgressEdgeTunnelFactory)
	if err := r.RegisterXrctrl(xgressEdgeTunnelFactory); err != nil {
		logrus.WithError(err).Panic("error registering edge tunnel in framework")
	}

	if err := r.RegisterXrctrl(r.GetStateManager()); err != nil {
		logrus.WithError(err).Panic("error registering state manager in framework")
	}

	if cliAgentEnabled {
		options := agent.Options{
			Addr:       cliAgentAddr,
			AppId:      config.Id.Token,
			AppType:    "router",
			AppVersion: version.GetVersion(),
			AppAlias:   cliAgentAlias,
		}
		if config.EnableDebugOps {
			enableDebugOps = true
		}
		r.RegisterDefaultAgentOps(enableDebugOps)
		debugops.RegisterEdgeRouterAgentOps(r, enableDebugOps)

		options.CustomOps = map[byte]func(conn net.Conn) error{
			agent.CustomOp:      r.HandleAgentOp,
			agent.CustomOpAsync: r.HandleAgentAsyncOp,
		}

		if err := agent.Listen(options); err != nil {
			pfxlog.Logger().WithError(err).Error("unable to start CLI agent")
		}
	}

	go waitForShutdown(r)

	if err := r.Run(); err != nil {
		logrus.WithError(err).Fatal("error starting")
	}
}

func getFlags(cmd *cobra.Command) map[string]*pflag.Flag {
	logtrace.LogWithFunctionName()
	ret := map[string]*pflag.Flag{}
	cmd.Flags().Visit(func(f *pflag.Flag) {
		ret[f.Name] = f
	})
	return ret
}

func waitForShutdown(r *router.Router) {
	logtrace.LogWithFunctionName()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)

	s := <-ch

	if s == syscall.SIGQUIT {
		fmt.Println("=== STACK DUMP BEGIN ===")
		debugz.DumpStack()
		fmt.Println("=== STACK DUMP CLOSE ===")
	}

	pfxlog.Logger().Info("shutting down ziti router")

	if err := r.Shutdown(); err != nil {
		pfxlog.Logger().WithError(err).Info("error encountered during shutdown")
	}
}
