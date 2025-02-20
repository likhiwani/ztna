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

package xgress_transport

import (
	"errors"
	"fmt"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/env"
	"ztna-core/ztna/router/xgress"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
)

type listener struct {
	id          *identity.TokenId
	ctrl        env.NetworkControllers
	options     *xgress.Options
	tcfg        transport.Configuration
	closeHelper *xgress.CloseHelper
}

func newListener(id *identity.TokenId, ctrl env.NetworkControllers, options *xgress.Options, tcfg transport.Configuration) xgress.Listener {
	logtrace.LogWithFunctionName()
	return &listener{
		id:          id,
		ctrl:        ctrl,
		options:     options,
		tcfg:        tcfg,
		closeHelper: &xgress.CloseHelper{},
	}
}

func (listener *listener) Listen(address string, bindHandler xgress.BindHandler) error {
	logtrace.LogWithFunctionName()
	if address == "" {
		return errors.New("address must be specified for transport listeners")
	}
	txAddress, err := transport.ParseAddress(address)
	if err != nil {
		return fmt.Errorf("cannot listen on invalid address [%s] (%s)", address, err)
	}

	acceptF := func(peer transport.Conn) {
		go listener.handleConnect(peer, bindHandler)
	}
	go listener.closeHelper.Init(txAddress.MustListen("tcp", listener.id, acceptF, listener.tcfg))

	return nil
}

func (listener *listener) Close() error {
	logtrace.LogWithFunctionName()
	return listener.closeHelper.Close()
}

func (listener *listener) handleConnect(peer transport.Conn, bindHandler xgress.BindHandler) {
	logtrace.LogWithFunctionName()
	conn := &transportXgressConn{Conn: peer}
	log := pfxlog.ContextLogger(conn.LogContext())

	request, err := xgress.ReceiveRequest(peer)
	if err == nil {
		response := xgress.CreateCircuit(listener.ctrl, conn, request, bindHandler, listener.options)
		err = xgress.SendResponse(response, peer)
		if err != nil {
			log.Errorf("error sending response (%s)", err)
		}

		if err != nil || !response.Success {
			if err := peer.Close(); err != nil {
				log.Errorf("error closing transport connection (%s)", err)
			}
		}
	} else {
		log.Errorf("error receiving request from peer (%s)", err)
		if err := peer.Close(); err != nil {
			log.Errorf("error closing transport connection (%s)", err)
		}
	}
}
