package tests

import (
	"math/big"
	"ztna-core/ztna/common/capabilities"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/config"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/versions"
	"github.com/openziti/transport/v2"
)

func (ctx *FabricTestContext) NewControlChannelListener() channel.UnderlayListener {
	logtrace.LogWithFunctionName()
	config, err := config.LoadConfig(FabricControllerConfFile)
	ctx.Req.NoError(err)
	ctx.Req.NoError(config.Db.Close())

	versionHeader, err := versions.StdVersionEncDec.Encode(versions.NewDefaultVersionProvider().AsVersionInfo())
	ctx.Req.NoError(err)

	capabilityMask := &big.Int{}
	capabilityMask.SetBit(capabilityMask, capabilities.ControllerCreateTerminatorV2, 1)
	capabilityMask.SetBit(capabilityMask, capabilities.ControllerSingleRouterLinkSource, 1)
	headers := map[int32][]byte{
		channel.HelloVersionHeader:                       versionHeader,
		int32(ctrl_pb.ControlHeaders_CapabilitiesHeader): capabilityMask.Bytes(),
	}

	ctrlChannelListenerConfig := channel.ListenerConfig{
		ConnectOptions:  config.Ctrl.Options.ConnectOptions,
		Headers:         headers,
		TransportConfig: transport.Configuration{"protocol": "ziti-ctrl"},
	}
	ctrlListener := channel.NewClassicListener(config.Id, config.Ctrl.Listener, ctrlChannelListenerConfig)
	ctx.Req.NoError(ctrlListener.Listen())
	return ctrlListener
}
