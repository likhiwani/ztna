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

package command

import (
	"reflect"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/debugz"
	"github.com/openziti/foundation/v2/rate"
	"github.com/openziti/storage/boltz"
	"github.com/sirupsen/logrus"
)

// Command instances represent actions to be taken by the fabric controller. They are serializable,
// so they can be shipped from one controller for RAFT coordination
type Command interface {
	// Apply runs the commands
	Apply(ctx boltz.MutateContext) error

	// GetChangeContext returns the change context associated with the command
	GetChangeContext() *change.Context

	// Encode returns a serialized representation of the command
	Encode() ([]byte, error)
}

// Validatable instances can be validated. Command instances which implement Validable will be validated
// before Command.Apply is called
type Validatable interface {
	Validate() error
}

// Dispatcher instances will take a command and either send it to the leader to be applied, or if the current
// system is the leader, apply it locally
type Dispatcher interface {
	Dispatch(command Command) error
	IsLeaderOrLeaderless() bool
	IsLeaderless() bool
	IsLeader() bool
	GetPeers() map[string]channel.Channel
	GetRateLimiter() rate.RateLimiter
	Bootstrap() error
}

// LocalDispatcher should be used when running a non-clustered system
type LocalDispatcher struct {
	EncodeDecodeCommands bool
	Limiter              rate.RateLimiter
}

func (self *LocalDispatcher) Bootstrap() error {
	logtrace.LogWithFunctionName()
	return nil
}

func (self *LocalDispatcher) IsLeader() bool {
	logtrace.LogWithFunctionName()
	return true
}

func (self *LocalDispatcher) IsLeaderOrLeaderless() bool {
	logtrace.LogWithFunctionName()
	return true
}

func (self *LocalDispatcher) IsLeaderless() bool {
	logtrace.LogWithFunctionName()
	return false
}

func (self *LocalDispatcher) GetPeers() map[string]channel.Channel {
	logtrace.LogWithFunctionName()
	return nil
}

func (self *LocalDispatcher) GetRateLimiter() rate.RateLimiter {
	logtrace.LogWithFunctionName()
	return self.Limiter
}

func (self *LocalDispatcher) Dispatch(command Command) error {
	logtrace.LogWithFunctionName()
	defer func() {
		if p := recover(); p != nil {
			pfxlog.Logger().
				WithField(logrus.ErrorKey, p).
				WithField("cmdType", reflect.TypeOf(command)).
				Error("error while dispatching command of type")
			debugz.DumpLocalStack()
			panic(p)
		}
	}()

	changeCtx := command.GetChangeContext()
	if changeCtx == nil {
		changeCtx = change.New().SetSourceType("unattributed").SetChangeAuthorType(change.AuthorTypeUnattributed)
	}

	if self.EncodeDecodeCommands {
		bytes, err := command.Encode()
		if err != nil {
			return err
		}
		cmd, err := GetDefaultDecoders().Decode(bytes)
		if err != nil {
			return err
		}
		command = cmd
	}

	return self.Limiter.RunRateLimited(func() error {
		ctx := changeCtx.NewMutateContext()
		return command.Apply(ctx)
	})
}

// Decoder instances know how to decode encoded commands
type Decoder interface {
	Decode(commandType int32, data []byte) (Command, error)
}

// DecoderF is a function version of the Decoder interface
type DecoderF func(commandType int32, data []byte) (Command, error)

func (self DecoderF) Decode(commandType int32, data []byte) (Command, error) {
	logtrace.LogWithFunctionName()
	return self(commandType, data)
}
