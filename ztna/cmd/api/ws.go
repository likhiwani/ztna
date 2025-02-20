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

package api

import (
	logtrace "ztna-core/ztna/logtrace"
	"io"
	"net/http"
	"strings"
	"time"

	"ztna-core/ztna/ztna/util"
	"github.com/gorilla/websocket"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/channel/v3/websockets"
	"github.com/openziti/identity"
)

func NewWsMgmtChannel(bindHandler channel.BindHandler) (channel.Channel, error) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger()
	restClientIdentity, err := util.LoadSelectedIdentity()
	if err != nil {
		return nil, err
	}

	baseUrl, err := restClientIdentity.GetBaseUrlForApi(util.FabricAPI)
	if err != nil {
		return nil, err
	}

	wsUrl := strings.ReplaceAll(baseUrl, "http", "ws") + "/ws-api"
	tlsConfig, err := restClientIdentity.NewTlsClientConfig()
	if err != nil {
		return nil, err
	}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		TLSClientConfig:  tlsConfig,
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(wsUrl, restClientIdentity.NewWsHeader())
	if err != nil {
		if resp != nil {
			if body, rerr := io.ReadAll(resp.Body); rerr == nil {
				log.WithError(err).Errorf("response body [%v]", string(body))
			}
		} else {
			log.WithError(err).Error("no response from websocket dial")
		}
		return nil, err
	}

	id := &identity.TokenId{Token: "mgmt"}
	underlayFactory := websockets.NewUnderlayFactory(id, conn, nil)

	ch, err := channel.NewChannel("mgmt", underlayFactory, bindHandler, nil)
	if err != nil {
		return nil, err
	}
	return ch, nil
}
