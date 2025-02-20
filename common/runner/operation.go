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

package runner

import (
	"time"
	"ztna-core/ztna/logtrace"

	"github.com/google/uuid"
)

type Operation interface {
	GetName() string
	GetId() uuid.UUID
	Run() error
	GetFrequency() time.Duration
	SetFrequency(duration time.Duration) error
}

type BaseOperation struct {
	Frequency time.Duration
	Id        uuid.UUID
	Name      string
}

func NewBaseOperation(name string, freq time.Duration) *BaseOperation {
	logtrace.LogWithFunctionName()
	return &BaseOperation{
		Name:      name,
		Id:        uuid.New(),
		Frequency: freq,
	}
}

func (e *BaseOperation) GetName() string {
	logtrace.LogWithFunctionName()
	return e.Name
}

func (e *BaseOperation) GetId() uuid.UUID {
	logtrace.LogWithFunctionName()
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}

	return e.Id
}

func (e *BaseOperation) GetFrequency() time.Duration {
	logtrace.LogWithFunctionName()
	return e.Frequency
}

func (e *BaseOperation) SetFrequency(f time.Duration) error {
	logtrace.LogWithFunctionName()
	e.Frequency = f
	return nil
}
