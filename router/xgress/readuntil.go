/*
	Copyright 2019 NetFoundry Inc.

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
	"errors"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/transport/v2"
)

func ReadUntilNewline(peer transport.Conn) ([]byte, error) {
	logtrace.LogWithFunctionName()
	return ReadUntil(peer, '\n')
}

func ReadUntil(peer transport.Conn, stop byte) ([]byte, error) {
	logtrace.LogWithFunctionName()
	buffer := make([]byte, 0)
	done := false
	for !done {
		next := make([]byte, 1)
		n, err := peer.Read(next)
		if err != nil {
			return nil, err
		}
		if n != 1 {
			return nil, errors.New("short read")
		}
		buffer = append(buffer, next[:n]...)
		done = buffer[len(buffer)-1] == stop
	}
	return buffer, nil
}
