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

package datastructures

import (
	logtrace "ztna-core/ztna/logtrace"

	"github.com/openziti/storage/objectz"
	cmap "github.com/orcaman/concurrent-map/v2"
)

func IterateCMap[T any](m cmap.ConcurrentMap[string, T]) objectz.ObjectIterator[T] {
	logtrace.LogWithFunctionName()
	iterator := &tupleChannelIterator[T]{
		c:     m.IterBuffered(),
		valid: true,
	}
	iterator.Next()
	return iterator
}

type tupleChannelIterator[T any] struct {
	c       <-chan cmap.Tuple[string, T]
	current T
	valid   bool
}

func (self *tupleChannelIterator[T]) IsValid() bool {
	logtrace.LogWithFunctionName()
	return self.valid
}

func (self *tupleChannelIterator[T]) Next() {
	logtrace.LogWithFunctionName()
	next, ok := <-self.c
	if !ok {
		self.valid = false
	} else {
		self.current = next.Val
	}
}

func (self *tupleChannelIterator[T]) Current() T {
	logtrace.LogWithFunctionName()
	return self.current
}
