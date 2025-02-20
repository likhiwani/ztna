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

package ctrl_msg

import (
	"errors"
	"fmt"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	logtrace "ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
)

const (
	CircuitSuccessType = 1001
	CircuitFailedType  = 1016
	RouteResultType    = 1022

	CircuitSuccessAddressHeader = 1100
	RouteResultAttemptHeader    = 1101
	RouteResultSuccessHeader    = 1102
	RouteResultErrorHeader      = 1103
	RouteResultErrorCodeHeader  = 1104

	TerminatorLocalAddressHeader  = 1110
	TerminatorRemoteAddressHeader = 1111

	InitiatorLocalAddressHeader  = 1112
	InitiatorRemoteAddressHeader = 1113

	XtStickinessToken = 1114

	ErrorTypeGeneric                 = 0
	ErrorTypeInvalidTerminator       = 1
	ErrorTypeMisconfiguredTerminator = 2
	ErrorTypeDialTimedOut            = 3
	ErrorTypeConnectionRefused       = 4

	CreateCircuitPeerDataHeader = 10

	CreateCircuitReqSessionTokenHeader         = 11
	CreateCircuitReqFingerprintsHeader         = 12
	CreateCircuitReqTerminatorInstanceIdHeader = 13
	CreateCircuitReqApiSessionTokenHeader      = 14

	CreateCircuitRespCircuitId  = 11
	CreateCircuitRespAddress    = 12
	CreateCircuitRespTagsHeader = 13

	HeaderResultErrorCode = 10

	ResultErrorRateLimited = 1
)

func NewCircuitSuccessMsg(sessionId, address string) *channel.Message {
	logtrace.LogWithFunctionName()
	msg := channel.NewMessage(CircuitSuccessType, []byte(sessionId))
	msg.Headers[CircuitSuccessAddressHeader] = []byte(address)
	return msg
}

func NewCircuitFailedMsg(message string) *channel.Message {
	logtrace.LogWithFunctionName()
	return channel.NewMessage(CircuitFailedType, []byte(message))
}

func NewRouteResultSuccessMsg(sessionId string, attempt int) *channel.Message {
	logtrace.LogWithFunctionName()
	msg := channel.NewMessage(RouteResultType, []byte(sessionId))
	msg.PutUint32Header(RouteResultAttemptHeader, uint32(attempt))
	msg.PutUint32Header(RouteResultAttemptHeader, uint32(attempt))
	msg.PutBoolHeader(RouteResultSuccessHeader, true)
	return msg
}

func NewRouteResultFailedMessage(sessionId string, attempt int, rerr string) *channel.Message {
	logtrace.LogWithFunctionName()
	msg := channel.NewMessage(RouteResultType, []byte(sessionId))
	msg.PutUint32Header(RouteResultAttemptHeader, uint32(attempt))
	msg.Headers[RouteResultErrorHeader] = []byte(rerr)
	return msg
}

type CreateCircuitRequest struct {
	ApiSessionToken      string
	SessionToken         string
	Fingerprints         []string
	TerminatorInstanceId string
	PeerData             map[uint32][]byte
}

func (self *CreateCircuitRequest) GetApiSessionToken() string {
	logtrace.LogWithFunctionName()
	return self.ApiSessionToken
}

func (self *CreateCircuitRequest) GetSessionToken() string {
	logtrace.LogWithFunctionName()
	return self.SessionToken
}

func (self *CreateCircuitRequest) GetFingerprints() []string {
	logtrace.LogWithFunctionName()
	return self.Fingerprints
}

func (self *CreateCircuitRequest) GetTerminatorInstanceId() string {
	logtrace.LogWithFunctionName()
	return self.TerminatorInstanceId
}

func (self *CreateCircuitRequest) GetPeerData() map[uint32][]byte {
	logtrace.LogWithFunctionName()
	return self.PeerData
}

func (self *CreateCircuitRequest) ToMessage() *channel.Message {
	logtrace.LogWithFunctionName()
	msg := channel.NewMessage(int32(edge_ctrl_pb.ContentType_CreateCircuitV2RequestType), nil)
	msg.PutStringHeader(CreateCircuitReqSessionTokenHeader, self.SessionToken)
	msg.PutStringHeader(CreateCircuitReqApiSessionTokenHeader, self.ApiSessionToken)
	msg.PutStringSliceHeader(CreateCircuitReqFingerprintsHeader, self.Fingerprints)
	msg.PutStringHeader(CreateCircuitReqTerminatorInstanceIdHeader, self.TerminatorInstanceId)
	msg.PutU32ToBytesMapHeader(CreateCircuitPeerDataHeader, self.PeerData)
	return msg
}

func DecodeCreateCircuitRequest(m *channel.Message) (*CreateCircuitRequest, error) {
	logtrace.LogWithFunctionName()
	sessionToken, _ := m.GetStringHeader(CreateCircuitReqSessionTokenHeader)
	if len(sessionToken) == 0 {
		return nil, errors.New("no session token provided in create circuit request")
	}

	apiSessionToken, _ := m.GetStringHeader(CreateCircuitReqApiSessionTokenHeader)

	fingerprints, _, err := m.GetStringSliceHeader(CreateCircuitReqFingerprintsHeader)
	if err != nil {
		return nil, fmt.Errorf("unable to get create circuit request fingerprints (%w)", err)
	}

	terminatorInstanceId, _ := m.GetStringHeader(CreateCircuitReqTerminatorInstanceIdHeader)
	peerData, _, err := m.GetU32ToBytesMapHeader(CreateCircuitPeerDataHeader)
	if err != nil {
		return nil, fmt.Errorf("unable to get create circuit request peer data (%w)", err)
	}

	return &CreateCircuitRequest{
		ApiSessionToken:      apiSessionToken,
		SessionToken:         sessionToken,
		Fingerprints:         fingerprints,
		TerminatorInstanceId: terminatorInstanceId,
		PeerData:             peerData,
	}, nil
}

type CreateCircuitResponse struct {
	CircuitId string
	Address   string
	PeerData  map[uint32][]byte
	Tags      map[string]string
}

func (self *CreateCircuitResponse) ToMessage() *channel.Message {
	logtrace.LogWithFunctionName()
	msg := channel.NewMessage(int32(edge_ctrl_pb.ContentType_CreateCircuitV2ResponseType), nil)
	msg.PutStringHeader(CreateCircuitRespCircuitId, self.CircuitId)
	msg.PutStringHeader(CreateCircuitRespAddress, self.Address)
	msg.PutU32ToBytesMapHeader(CreateCircuitPeerDataHeader, self.PeerData)
	msg.PutStringToStringMapHeader(CreateCircuitRespTagsHeader, self.Tags)
	return msg
}

func DecodeCreateCircuitResponse(m *channel.Message) (*CreateCircuitResponse, error) {
	logtrace.LogWithFunctionName()
	circuitId, _ := m.GetStringHeader(CreateCircuitRespCircuitId)
	address, _ := m.GetStringHeader(CreateCircuitRespAddress)
	peerData, _, err := m.GetU32ToBytesMapHeader(CreateCircuitPeerDataHeader)
	if err != nil {
		return nil, fmt.Errorf("unable to get create circuit response peer data (%w)", err)
	}

	tags, _, err := m.GetStringToStringMapHeader(CreateCircuitRespTagsHeader)
	if err != nil {
		return nil, fmt.Errorf("unable to get create circuit response tags (%w)", err)
	}

	return &CreateCircuitResponse{
		CircuitId: circuitId,
		Address:   address,
		PeerData:  peerData,
		Tags:      tags,
	}, nil
}
