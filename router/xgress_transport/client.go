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
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/xgress"

	"github.com/openziti/identity"
	"github.com/openziti/transport/v2"
)

// ClientDial dials the given xgress address and handles authentication, returning an authed connection or an error
func ClientDial(addr transport.Address, id *identity.TokenId, serviceId *identity.TokenId, tcfg transport.Configuration) (transport.Conn, error) {
	logtrace.LogWithFunctionName()
	peer, err := addr.Dial("i/"+id.Token, id, 0, tcfg)
	if err != nil {
		return nil, err
	}

	request := &xgress.Request{
		Id:        id.Token,
		ServiceId: serviceId.Token,
	}
	err = xgress.SendRequest(request, peer)
	if err != nil {
		return nil, err
	}
	response, err := xgress.ReceiveResponse(peer)
	if err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, errors.New(response.Message)
	}

	return peer, nil
}
