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

package xgress

import (
	"fmt"
	"ztna-core/ztna/logtrace"
)

type registry struct {
	factories map[string]Factory
}

func NewRegistry() *registry {
	logtrace.LogWithFunctionName()
	return &registry{
		factories: make(map[string]Factory),
	}
}

func (registry *registry) Register(name string, factory Factory) {
	logtrace.LogWithFunctionName()
	registry.factories[name] = factory
}

func (registry *registry) Factory(name string) (Factory, error) {
	logtrace.LogWithFunctionName()
	if factory, found := registry.factories[name]; found {
		return factory, nil
	} else {
		return nil, fmt.Errorf("xgress factory [%s] not found", name)
	}
}

func (registry *registry) List() []string {
	logtrace.LogWithFunctionName()
	names := make([]string, 0)
	for k := range registry.factories {
		names = append(names, k)
	}
	return names
}

func (registry *registry) Debug() string {
	logtrace.LogWithFunctionName()
	out := "{"
	for _, name := range registry.List() {
		out += " " + name
	}
	out += " }"
	return out
}

var globalRegistry *registry

func GlobalRegistry() *registry {
	logtrace.LogWithFunctionName()
	if globalRegistry == nil {
		globalRegistry = NewRegistry()
	}
	return globalRegistry
}
