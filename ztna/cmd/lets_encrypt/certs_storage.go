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
	"bytes"
	"crypto/x509"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ztna-core/ztna/ztna/internal/log"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"golang.org/x/net/idna"
)

const filePerm os.FileMode = 0o600

const (
	baseCertificatesFolderName = "certificates"
	baseArchivesFolderName     = "archives"
)

// CertificatesStorage a certificates storage.
//
// rootPath:
//
//	./.path/certificates/
//	     │      └── root certificates directory
//	     └── "path" option
//
// archivePath:
//
//	./.path/archives/
//	     │      └── archived certificates directory
//	     └── "path" option
type CertificatesStorage struct {
	rootPath    string
	archivePath string
	pem         bool
}

// NewCertificatesStorage create a new certificates storage.
func NewCertificatesStorage(path string) *CertificatesStorage {
	abspath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	return &CertificatesStorage{
		rootPath:    filepath.Join(abspath, baseCertificatesFolderName),
		archivePath: filepath.Join(abspath, baseArchivesFolderName),
		pem:         true,
	}
}

func (s *CertificatesStorage) CreateRootFolder() {
	err := createNonExistingFolder(s.rootPath)
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}
}

func (s *CertificatesStorage) CreateArchiveFolder() {
	err := createNonExistingFolder(s.archivePath)
	if err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}
}

func (s *CertificatesStorage) GetRootPath() string {
	return s.rootPath
}

func (s *CertificatesStorage) SaveResource(certRes *certificate.Resource) {
	domain := certRes.Domain

	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	err := s.WriteFile(domain, ".crt", certRes.Certificate)
	if err != nil {
		log.Fatalf("Unable to save Certificate for domain %s\n\t%v", domain, err)
	}

	if certRes.IssuerCertificate != nil {
		err = s.WriteFile(domain, ".issuer.crt", certRes.IssuerCertificate)
		if err != nil {
			log.Fatalf("Unable to save IssuerCertificate for domain %s\n\t%v", domain, err)
		}
	}

	if certRes.PrivateKey != nil {
		// if we were given a CSR, we don't know the private key
		err = s.WriteFile(domain, ".key", certRes.PrivateKey)
		if err != nil {
			log.Fatalf("Unable to save PrivateKey for domain %s\n\t%v", domain, err)
		}

		if s.pem {
			err = s.WriteFile(domain, ".pem", bytes.Join([][]byte{certRes.Certificate, certRes.PrivateKey}, nil))
			if err != nil {
				log.Fatalf("Unable to save Certificate and PrivateKey in .pem for domain %s\n\t%v", domain, err)
			}
		}
	} else if s.pem {
		// we don't have the private key; can't write the .pem file
		log.Fatalf("Unable to save pem without private key for domain %s\n\t%v; are you using a CSR?", domain, err)
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		log.Fatalf("Unable to marshal CertResource for domain %s\n\t%v", domain, err)
	}

	err = s.WriteFile(domain, ".json", jsonBytes)
	if err != nil {
		log.Fatalf("Unable to save CertResource for domain %s\n\t%v", domain, err)
	}

	log.Infof("certs successfully created in %v", s.rootPath)
}

func (s *CertificatesStorage) ReadResource(domain string) certificate.Resource {
	raw, err := s.ReadFile(domain, ".json")
	if err != nil {
		log.Fatalf("Error while loading the meta data for domain %s\n\t%v", domain, err)
	}

	var resource certificate.Resource
	if err = json.Unmarshal(raw, &resource); err != nil {
		log.Fatalf("Error while marshaling the meta data for domain %s\n\t%v", domain, err)
	}

	return resource
}

func (s *CertificatesStorage) ExistsFile(domain, extension string) bool {
	filePath := s.GetFileName(domain, extension)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Fatalf("%v", err)
	}
	return true
}

func (s *CertificatesStorage) ReadFile(domain, extension string) ([]byte, error) {
	return os.ReadFile(s.GetFileName(domain, extension))
}

func (s *CertificatesStorage) GetFileName(domain, extension string) string {
	filename := sanitizedDomain(domain) + extension
	return filepath.Join(s.rootPath, filename)
}

func (s *CertificatesStorage) ReadCertificate(domain, extension string) ([]*x509.Certificate, error) {
	content, err := s.ReadFile(domain, extension)
	if err != nil {
		return nil, err
	}

	// The input may be a bundle or a single certificate.
	return certcrypto.ParsePEMBundle(content)
}

func (s *CertificatesStorage) WriteFile(domain, extension string, data []byte) error {
	baseFileName := sanitizedDomain(domain)
	filePath := filepath.Join(s.rootPath, baseFileName+extension)
	return os.WriteFile(filePath, data, filePerm)
}

func (s *CertificatesStorage) MoveToArchive(domain string) error {
	matches, err := filepath.Glob(filepath.Join(s.rootPath, sanitizedDomain(domain)+".*"))
	if err != nil {
		return err
	}

	for _, oldFile := range matches {
		date := strconv.FormatInt(time.Now().Unix(), 10)
		filename := date + "." + filepath.Base(oldFile)
		newFile := filepath.Join(s.archivePath, filename)

		err = os.Rename(oldFile, newFile)
		if err != nil {
			return err
		}
	}

	return nil
}

// sanitizedDomain Make sure no funny chars are in the cert names (like wildcards ;)).
func sanitizedDomain(domain string) string {
	safe, err := idna.ToASCII(strings.ReplaceAll(domain, "*", "_"))
	if err != nil {
		log.Fatalf("%v", err)
	}
	return safe
}
