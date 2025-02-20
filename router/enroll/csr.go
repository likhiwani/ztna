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

package enroll

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/internal/edgerouter"
)

type Csr struct {
	Pem      string
	Name     string
	SanDns   []string
	SanEmail []string
	SanIp    []string
	SanUri   []string
}

func CreateCsr(key crypto.PrivateKey, algo x509.SignatureAlgorithm, subj *pkix.Name, sans *edgerouter.Sans) (string, error) {
	logtrace.LogWithFunctionName()
	template := x509.CertificateRequest{
		Subject:            *subj,
		SignatureAlgorithm: algo,
		DNSNames:           sans.DnsAddresses,
		EmailAddresses:     sans.EmailAddresses,
		IPAddresses:        sans.IpAddressesParsed,
		URIs:               sans.UriAddressesParsed,
	}
	csrBytes, csrErr := x509.CreateCertificateRequest(rand.Reader, &template, key)

	if csrErr != nil {
		return "", csrErr
	}

	outBuff := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	return string(outBuff), nil
}
