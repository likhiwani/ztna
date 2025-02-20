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

package handler_peer_ctrl

import (
	"time"
	"ztna-core/ztna/common/metrics"
	"ztna-core/ztna/common/pb/cmd_pb"
	"ztna-core/ztna/controller/apierror"
	"ztna-core/ztna/controller/raft"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/goroutines"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func newCommandHandler(controller *raft.Controller) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	poolConfig := goroutines.PoolConfig{
		QueueSize:   uint32(controller.Config.CommandHandlerOptions.MaxQueueSize),
		MinWorkers:  0,
		MaxWorkers:  1, // we should only have one thing apply entries, so they don't get applied out of order
		IdleTime:    time.Second,
		CloseNotify: controller.GetCloseNotify(),
		PanicHandler: func(err interface{}) {
			pfxlog.Logger().WithField(logrus.ErrorKey, err).Error("panic during command processing")
		},
	}
	metrics.ConfigureGoroutinesPoolMetrics(&poolConfig, controller.GetMetricsRegistry(), "command_handler")
	pool, err := goroutines.NewPool(poolConfig)
	if err != nil {
		panic(err)
	}
	return &commandHandler{
		controller: controller,
		pool:       pool,
	}
}

type commandHandler struct {
	controller *raft.Controller
	pool       goroutines.Pool
}

func (self *commandHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(cmd_pb.ContentType_NewLogEntryType)
}

func (self *commandHandler) HandleReceive(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label())

	err := self.pool.QueueOrError(func() {
		if idx, err := self.controller.ApplyEncodedCommand(m.Body); err != nil {
			sendErrorResponseCalculateType(m, ch, err)
			return
		} else {
			sendSuccessResponse(m, ch, idx)
		}
	})

	if errors.Is(err, goroutines.QueueFullError) {
		err = apierror.NewTooManyUpdatesError()
	}

	if err != nil {
		log.WithError(err).Error("unable to queue command for processing")
		go sendErrorResponseCalculateType(m, ch, err)
	}
}
