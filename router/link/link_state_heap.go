/*
	(c) Copyright NetFoundry Inc.

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

package link

import "ztna-core/ztna/logtrace"

type linkStateHeap []*linkState

func (self linkStateHeap) Len() int {
	logtrace.LogWithFunctionName()
	return len(self)
}

func (self linkStateHeap) Less(i, j int) bool {
	logtrace.LogWithFunctionName()
	return self[i].nextDial.Before(self[j].nextDial)
}

func (self linkStateHeap) Swap(i, j int) {
	logtrace.LogWithFunctionName()
	tmp := self[i]
	self[i] = self[j]
	self[j] = tmp
}

func (self *linkStateHeap) Push(x any) {
	logtrace.LogWithFunctionName()
	*self = append(*self, x.(*linkState))
}

func (self *linkStateHeap) Pop() any {
	logtrace.LogWithFunctionName()
	old := *self
	n := len(old)
	pm := old[n-1]
	*self = old[0 : n-1]
	return pm
}
