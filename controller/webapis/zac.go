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

package webapis

import (
	"fmt"
	"strings"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/xweb/v2"
	log "github.com/sirupsen/logrus"
)

const (
	Binding = "zac"
)

type ZitiAdminConsoleFactory struct {
}

var _ xweb.ApiHandlerFactory = &ZitiAdminConsoleFactory{}

func NewZitiAdminConsoleFactory() *ZitiAdminConsoleFactory {
	logtrace.LogWithFunctionName()
	return &ZitiAdminConsoleFactory{}
}

func (factory *ZitiAdminConsoleFactory) Validate(*xweb.InstanceConfig) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (factory *ZitiAdminConsoleFactory) Binding() string {
	logtrace.LogWithFunctionName()
	return Binding
}

func (factory *ZitiAdminConsoleFactory) New(_ *xweb.ServerConfig, options map[interface{}]interface{}) (xweb.ApiHandler, error) {
	logtrace.LogWithFunctionName()
	locVal := options["location"]
	if locVal == nil || locVal == "" {
		return nil, fmt.Errorf("location must be supplied in the %s options", Binding)
	}

	loc, ok := locVal.(string)

	if !ok {
		return nil, fmt.Errorf("location must be a string for the %s options", Binding)
	}

	indexFileVal := options["indexFile"]
	indexFile := "index.html"

	if indexFileVal != nil {
		newFileVal, ok := indexFileVal.(string)

		if !ok {
			return nil, fmt.Errorf("indexFile must be a string for the %s options", Binding)
		}

		newFileVal = strings.TrimSpace(newFileVal)

		if newFileVal != "" {
			indexFile = newFileVal
		}
	}

	contextRoot := "/" + Binding
	spa := &GenericHttpHandler{
		HttpHandler: SpaHandler(loc, contextRoot, indexFile),
		BindingKey:  Binding,
		ContextRoot: contextRoot,
	}

	log.Infof("initializing ZAC SPA Handler from %s", locVal)
	return spa, nil
}
