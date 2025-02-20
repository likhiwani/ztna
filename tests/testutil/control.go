package testutil

import (
	"time"
	"ztna-core/ztna/common/handler_common"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/stretchr/testify/require"
)

func AcceptControl(id string, uf channel.UnderlayFactory, assertions *require.Assertions) (channel.Channel, *MessageCollector) {
	logtrace.LogWithFunctionName()
	msgc := NewMessageCollector(id)
	bindHandler := func(binding channel.Binding) error {
		binding.AddReceiveHandler(channel.AnyContentType, msgc)
		binding.AddReceiveHandlerF(int32(ctrl_pb.ContentType_VerifyRouterType), func(msg *channel.Message, ch channel.Channel) {
			handler_common.SendSuccess(msg, ch, "link success")
		})
		return nil
	}

	timeoutUF := NewTimeoutUnderlayFactory(uf, 2*time.Second)
	ch, err := channel.NewChannel(id, timeoutUF, channel.BindHandlerF(bindHandler), channel.DefaultOptions())
	assertions.NoError(err)
	return ch, msgc
}
