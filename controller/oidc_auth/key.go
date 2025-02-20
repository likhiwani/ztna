package oidc_auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"ztna-core/ztna/logtrace"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/go-jose/go-jose.v2"
)

// key implements op.Key and represents a private key
type key struct {
	id         string
	algorithm  jose.SignatureAlgorithm
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
}

func (p *key) Algorithm() jose.SignatureAlgorithm {
	logtrace.LogWithFunctionName()
	return p.algorithm
}

func (p *key) Use() string {
	logtrace.LogWithFunctionName()
	return "sig"
}

func (s *key) SignatureAlgorithm() jose.SignatureAlgorithm {
	logtrace.LogWithFunctionName()
	return s.algorithm
}

// Key returns the private key for the key pair.
func (s *key) Key() interface{} {
	logtrace.LogWithFunctionName()
	return s.privateKey
}

func (s *key) ID() string {
	logtrace.LogWithFunctionName()
	return s.id
}

// pubKey implements op.Key and represents a public key
type pubKey struct {
	key
}

// Key returns the public key for the key pair.
func (s *pubKey) Key() interface{} {
	logtrace.LogWithFunctionName()
	return s.publicKey
}

// newKeyFromCert will create a new PubKey from an x509.Certificate
func newKeyFromCert(cert *x509.Certificate, kid string) *pubKey {
	logtrace.LogWithFunctionName()
	signingMethod := getSigningMethod(cert)

	if signingMethod == nil {
		return nil
	}

	return &pubKey{
		key{
			id:        kid,
			algorithm: jose.SignatureAlgorithm(signingMethod.Alg()),
			publicKey: cert.PublicKey,
		},
	}
}

// getSigningMethod determines the jwt.SigningMethod necessary for certificate
func getSigningMethod(cert *x509.Certificate) jwt.SigningMethod {
	logtrace.LogWithFunctionName()
	switch pubKey := cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		switch pubKey.Params().BitSize {
		case jwt.SigningMethodES256.CurveBits:
			return jwt.SigningMethodES256
		case jwt.SigningMethodES384.CurveBits:
			return jwt.SigningMethodES384
		case jwt.SigningMethodES512.CurveBits:
			return jwt.SigningMethodES512
		}
	case *rsa.PublicKey:
		return jwt.SigningMethodRS256
	}

	return nil
}
