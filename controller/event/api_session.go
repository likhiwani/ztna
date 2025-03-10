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

package event

import (
	"fmt"
	"time"
)

const ApiSessionEventTypeCreated = "created"
const ApiSessionEventTypeDeleted = "deleted"
const ApiSessionEventTypeRefreshed = "refreshed"
const ApiSessionEventTypeExchanged = "exchanged"
const ApiSessionEventNS = "edge.apiSessions"

const ApiSessionTypeLegacy = "legacy"
const ApiSessionTypeJwt = "jwt"

type ApiSessionEvent struct {
	Namespace  string    `json:"namespace"`
	EventType  string    `json:"event_type"`
	EventSrcId string    `json:"event_src_id"`
	Id         string    `json:"id"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
	Token      string    `json:"token"`
	IdentityId string    `json:"identity_id"`
	IpAddress  string    `json:"ip_address"`
}

func (event *ApiSessionEvent) String() string {
	return fmt.Sprintf("%v.%v id=%v timestamp=%v token=%v identityId=%v ipAddress=%v",
		event.Namespace, event.EventType, event.Id, event.Timestamp, event.Token, event.IdentityId, event.IpAddress)
}

type ApiSessionEventHandler interface {
	AcceptApiSessionEvent(event *ApiSessionEvent)
}

type ApiSessionEventHandlerWrapper interface {
	ApiSessionEventHandler
	IsWrapping(value ApiSessionEventHandler) bool
}
