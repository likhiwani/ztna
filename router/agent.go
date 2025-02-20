package router

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"time"
	"ztna-core/ztna/common/handler_common"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/common/pb/mgmt_pb"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	AgentAppId byte = 2
)

func (self *Router) RegisterAgentBindHandler(bindHandler channel.BindHandler) {
	logtrace.LogWithFunctionName()
	self.agentBindHandlers = append(self.agentBindHandlers, bindHandler)
}

func (self *Router) RegisterDefaultAgentOps(debugEnabled bool) {
	logtrace.LogWithFunctionName()
	self.agentBindHandlers = append(self.agentBindHandlers, channel.BindHandlerF(func(binding channel.Binding) error {
		binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_RouterDebugDumpForwarderTablesRequestType), self.agentOpDumpForwarderTables)
		binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_RouterDebugDumpLinksRequestType), self.agentOpsDumpLinks)
		binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_RouterQuiesceRequestType), self.agentOpQuiesceRouter)
		binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_RouterDequiesceRequestType), self.agentOpDequiesceRouter)
		binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_RouterDecommissionRequestType), self.agentOpDecommissionRouter)

		if debugEnabled {
			binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_RouterDebugUpdateRouteRequestType), self.agentOpUpdateRoute)
			binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_RouterDebugUnrouteRequestType), self.agentOpUnroute)
			binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_RouterDebugForgetLinkRequestType), self.agentOpForgetLink)
			binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_RouterDebugToggleCtrlChannelRequestType), self.agentOpToggleCtrlChan)
		}
		return nil
	}))
}

func (self *Router) RegisterAgentOp(opId byte, f func(c *bufio.ReadWriter) error) {
	logtrace.LogWithFunctionName()
	self.debugOperations[opId] = f
}

func (self *Router) bindAgentChannel(binding channel.Binding) error {
	logtrace.LogWithFunctionName()
	for _, bh := range self.agentBindHandlers {
		if err := binding.Bind(bh); err != nil {
			return err
		}
	}
	return nil
}

func (self *Router) HandleAgentAsyncOp(conn net.Conn) error {
	logtrace.LogWithFunctionName()
	logrus.Debug("received agent operation request")

	appIdBuf := []byte{0}
	_, err := io.ReadFull(conn, appIdBuf)
	if err != nil {
		return err
	}
	appId := appIdBuf[0]

	if appId != AgentAppId {
		logrus.WithField("appId", appId).Debug("invalid app id on agent request")
		return errors.New("invalid operation for controller")
	}

	options := channel.DefaultOptions()
	options.ConnectTimeout = time.Second
	listener := channel.NewExistingConnListener(self.config.Id, conn, nil)
	_, err = channel.NewChannel("agent", listener, channel.BindHandlerF(self.bindAgentChannel), options)
	return err
}

func (self *Router) agentOpForgetLink(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger()
	linkId := string(m.Body)
	var found bool
	if link, _ := self.xlinkRegistry.GetLinkById(linkId); link != nil {
		self.xlinkRegistry.DebugForgetLink(linkId)
		self.forwarder.UnregisterLink(link)
		found = true
	}

	log.Infof("forget of link %v was requested. link found? %v", linkId, found)
	result := fmt.Sprintf("link removed: %v", found)
	handler_common.SendOpResult(m, ch, "link.remove", result, true)
}

func (self *Router) agentOpToggleCtrlChan(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	ctrlId := string(m.Body)

	results := &bytes.Buffer{}
	toggleOn, _ := m.GetBoolHeader(int32(mgmt_pb.Header_CtrlChanToggle))

	success := true
	count := 0
	self.ctrls.ForEach(func(controllerId string, ch channel.Channel) {
		if ctrlId == "" || controllerId == ctrlId {
			log := pfxlog.Logger().WithField("ctrlId", controllerId)
			if toggleable, ok := ch.Underlay().(connectionToggle); ok {
				if toggleOn {
					if err := toggleable.Reconnect(); err != nil {
						log.WithError(err).Error("control channel: failed to reconnect")
						_, _ = fmt.Fprintf(results, "control channel: failed to reconnect (%v)\n", err)
						success = false
					} else {
						log.Warn("control channel: reconnected")
						_, _ = fmt.Fprint(results, "control channel: reconnected")
						count++
					}
				} else {
					if err := toggleable.Disconnect(); err != nil {
						log.WithError(err).Error("control channel: failed to close")
						_, _ = fmt.Fprintf(results, "control channel: failed to close (%v)\n", err)
						success = false
					} else {
						log.Warn("control channel: closed")
						_, _ = fmt.Fprint(results, "control channel: closed")
						count++
					}
				}
			} else {
				log.Warn("control channel: not toggleable")
				_, _ = fmt.Fprint(results, "control channel: not toggleable")
				success = false
			}
		}
	})

	if count == 0 {
		_, _ = fmt.Fprintf(results, "control channel: no controllers matched id [%v]", ctrlId)
		success = false
	}

	handler_common.SendOpResult(m, ch, "ctrl.toggle", results.String(), success)
}

func (self *Router) agentOpQuiesceRouter(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	ctrlCh := self.ctrls.AnyValidCtrlChannel()
	if ctrlCh == nil {
		handler_common.SendOpResult(m, ch, "quiesce", "unable to reach controller", false)
		return
	}

	msg := channel.NewMessage(int32(ctrl_pb.ContentType_QuiesceRouterRequestType), nil)
	resp, err := msg.WithTimeout(5 * time.Second).SendForReply(ctrlCh)
	if err != nil {
		handler_common.SendOpResult(m, ch, "quiesce", fmt.Sprintf("error in controller communications: %v", err.Error()), false)
		return
	}

	result := channel.UnmarshalResult(resp)
	handler_common.SendOpResult(m, ch, "quiesce", result.Message+"\n", result.Success)
}

func (self *Router) agentOpDequiesceRouter(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	ctrlCh := self.ctrls.AnyValidCtrlChannel()
	if ctrlCh == nil {
		handler_common.SendOpResult(m, ch, "dequiesce", "unable to reach controller", false)
		return
	}

	msg := channel.NewMessage(int32(ctrl_pb.ContentType_DequiesceRouterRequestType), nil)
	resp, err := msg.WithTimeout(5 * time.Second).SendForReply(ctrlCh)
	if err != nil {
		handler_common.SendOpResult(m, ch, "dequiesce", fmt.Sprintf("error in controller communications: %v", err.Error()), false)
		return
	}

	result := channel.UnmarshalResult(resp)
	handler_common.SendOpResult(m, ch, "dequiesce", result.Message+"\n", result.Success)
}

func (self *Router) agentOpDecommissionRouter(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	ctrlCh := self.ctrls.AnyValidCtrlChannel()
	if ctrlCh == nil {
		handler_common.SendOpResult(m, ch, "decomission", "unable to reach controller", false)
		return
	}

	msg := channel.NewMessage(int32(ctrl_pb.ContentType_DecommissionRouterRequestType), nil)
	err := msg.WithTimeout(5 * time.Second).SendAndWaitForWire(ctrlCh)
	if err != nil {
		handler_common.SendOpResult(m, ch, "decommission", "unexpected result, no disconnect but no error", false)
		return
	}

	connectable, ok := ctrlCh.Underlay().(interface{ IsConnected() bool })
	if !ok {
		pfxlog.Logger().Warn("control channel can't be checked to see if it's connected")
	}

	for i := 0; i < 100; i++ {
		// decommissioning router should cause control channel to close
		if connectable == nil || !connectable.IsConnected() {
			handler_common.SendOpResult(m, ch, "decommission", "router decommissioned\n", true)
			time.Sleep(500 * time.Millisecond)
			pfxlog.Logger().Info("router decommissioned, shutting down")
			os.Exit(0)
		}
		time.Sleep(10 * time.Millisecond)
	}

	pfxlog.Logger().Warn("control channel still connected after decomission")
	handler_common.SendOpResult(m, ch, "decommission", "controller didn't disconnect after decommission", false)
}

func (self *Router) agentOpDumpForwarderTables(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	tables := self.forwarder.Debug()
	handler_common.SendOpResult(m, ch, "dump.forwarder_tables", tables, true)
}

func (self *Router) agentOpsDumpLinks(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	result := &bytes.Buffer{}
	for link := range self.xlinkRegistry.Iter() {
		line := fmt.Sprintf("id: %v dest: %v protocol: %v\n", link.Id(), link.DestinationId(), link.LinkProtocol())
		_, err := result.WriteString(line)
		if err != nil {
			handler_common.SendOpResult(m, ch, "dump.links", err.Error(), false)
			return
		}
	}

	output := result.String()
	if len(output) == 0 {
		output = "no links\n"
	}
	handler_common.SendOpResult(m, ch, "dump.links", output, true)
}

func (self *Router) agentOpUpdateRoute(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	logrus.Warn("received debug operation to update routes")
	ctrlId, _ := m.GetStringHeader(int32(mgmt_pb.Header_ControllerId))
	if ctrlId == "" {
		handler_common.SendOpResult(m, ch, "update.route", "no controller id provided", false)
		return
	}

	ctrl := self.ctrls.GetCtrlChannel(ctrlId)
	if ctrl == nil {
		handler_common.SendOpResult(m, ch, "update.route", fmt.Sprintf("no control channel found for [%v]", ctrlId), false)
		return
	}

	route := &ctrl_pb.Route{}
	if err := proto.Unmarshal(m.Body, route); err != nil {
		handler_common.SendOpResult(m, ch, "update.route", err.Error(), false)
		return
	}

	if err := self.forwarder.Route(ctrlId, route); err != nil {
		handler_common.SendOpResult(m, ch, "update.route", errors.Wrap(err, "error adding route").Error(), false)
		return
	}

	logrus.Warnf("route added: %+v", route)
	handler_common.SendOpResult(m, ch, "update.route", "route added", true)
}

func (self *Router) agentOpUnroute(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	logrus.Warn("received debug operation to unroute a circuit")

	unroute := &ctrl_pb.Unroute{}
	if err := proto.Unmarshal(m.Body, unroute); err != nil {
		handler_common.SendOpResult(m, ch, "unroute", err.Error(), false)
		return
	}

	self.forwarder.Unroute(unroute.CircuitId, unroute.Now)
	logrus.Warnf("circuit unrouted: %+v", unroute)
	handler_common.SendOpResult(m, ch, "unroute", fmt.Sprintf("circuit unrouted [%s]", unroute.CircuitId), true)
}

func (self *Router) HandleAgentOp(conn net.Conn) error {
	logtrace.LogWithFunctionName()
	bconn := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	appId, err := bconn.ReadByte()
	if err != nil {
		return err
	}

	if appId != AgentAppId {
		return errors.Errorf("invalid operation for router")
	}

	op, err := bconn.ReadByte()

	if err != nil {
		return err
	}

	if opF, ok := self.debugOperations[op]; ok {
		if err := opF(bconn); err != nil {
			return err
		}
		return bconn.Flush()
	}
	return errors.Errorf("invalid operation %v", op)
}
