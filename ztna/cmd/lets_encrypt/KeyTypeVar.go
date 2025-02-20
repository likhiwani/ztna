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

package lets_encrypt

import (
	logtrace "ztna-core/ztna/logtrace"
	"errors"
	"github.com/go-acme/lego/v4/certcrypto"
)

type KeyTypeVar string

func (f *KeyTypeVar) String() string {
	logtrace.LogWithFunctionName()
	switch f.Get() {
	case certcrypto.EC256:
		return "EC256"
	case certcrypto.EC384:
		return "EC384"
	case certcrypto.RSA2048:
		return "RSA2048"
	case certcrypto.RSA4096:
		return "RSA4096"
	case certcrypto.RSA8192:
		return "RSA8192"
	default:
		return "?"
	}
}

func (f *KeyTypeVar) Set(value string) error {
	logtrace.LogWithFunctionName()
	switch value {
	case "EC256":
		*f = KeyTypeVar(certcrypto.EC256)
	case "EC384":
		*f = KeyTypeVar(certcrypto.EC384)
	case "RSA2048":
		*f = KeyTypeVar(certcrypto.RSA2048)
	case "RSA4096":
		*f = KeyTypeVar(certcrypto.RSA4096)
	case "RSA8192":
		*f = KeyTypeVar(certcrypto.RSA8192)
	default:
		return errors.New("Invalid option")
	}

	return nil
}

func (f *KeyTypeVar) EC256() bool {
	logtrace.LogWithFunctionName()
	return certcrypto.KeyType(*f) == certcrypto.EC256
}

func (f *KeyTypeVar) EC384() bool {
	logtrace.LogWithFunctionName()
	return certcrypto.KeyType(*f) == certcrypto.EC384
}

func (f *KeyTypeVar) RSA2048() bool {
	logtrace.LogWithFunctionName()
	return certcrypto.KeyType(*f) == certcrypto.RSA2048
}

func (f *KeyTypeVar) RSA4096() bool {
	logtrace.LogWithFunctionName()
	return certcrypto.KeyType(*f) == certcrypto.RSA4096
}

func (f *KeyTypeVar) RSA8192() bool {
	logtrace.LogWithFunctionName()
	return certcrypto.KeyType(*f) == certcrypto.RSA8192
}

func (f *KeyTypeVar) Get() certcrypto.KeyType {
	logtrace.LogWithFunctionName()
	return certcrypto.KeyType(*f)
}

func (f *KeyTypeVar) Type() string {
	logtrace.LogWithFunctionName()
	return "EC256|EC384|RSA2048|RSA4096|RSA8192"
}
