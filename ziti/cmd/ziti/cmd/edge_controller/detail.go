/*
	Copyright NetFoundry, Inc.

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

package edge_controller

import (
	"github.com/Jeffail/gabs"
	"github.com/openziti/ziti/ziti/cmd/ziti/util"
	"io"
)

func DetailEntityOfType(entityType, entityId string, logJSON bool, out io.Writer) (*gabs.Container, error) {
	jsonParsed, err := util.EdgeControllerDetailEntity(entityType, entityId, logJSON, out)

	if err != nil {
		return nil, err
	}

	entity := jsonParsed.S("data")
	return entity, nil
}
