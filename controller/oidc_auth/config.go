package oidc_auth

import (
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"time"
	"ztna-core/ztna/common"
	"ztna-core/ztna/logtrace"
)

// Config represents the configuration necessary to operate an OIDC Provider
type Config struct {
	Issuers              []string
	TokenSecret          string
	Storage              Storage
	Certificate          *x509.Certificate
	PrivateKey           crypto.PrivateKey
	IdTokenDuration      time.Duration
	RefreshTokenDuration time.Duration
	AccessTokenDuration  time.Duration
	RedirectURIs         []string
	PostLogoutURIs       []string

	maxTokenDuration *time.Duration
}

// NewConfig will create a Config with default values
func NewConfig(issuers []string, cert *x509.Certificate, key crypto.PrivateKey) Config {
	logtrace.LogWithFunctionName()
	return Config{
		Issuers:              issuers,
		Certificate:          cert,
		PrivateKey:           key,
		RefreshTokenDuration: common.DefaultRefreshTokenDuration,
		AccessTokenDuration:  common.DefaultAccessTokenDuration,
		IdTokenDuration:      common.DefaultIdTokenDuration,
	}
}

// MaxTokenDuration returns the maximum token lifetime currently configured
func (c *Config) MaxTokenDuration() time.Duration {
	logtrace.LogWithFunctionName()
	if c.maxTokenDuration == nil {
		curMaxDur := c.RefreshTokenDuration

		for _, duration := range []time.Duration{c.AccessTokenDuration, c.IdTokenDuration} {
			if duration > curMaxDur {
				curMaxDur = duration
			}
		}

		c.maxTokenDuration = &curMaxDur
	}

	return *c.maxTokenDuration
}

// Secret returns a sha256 sum of the configured token secret
func (c *Config) Secret() [32]byte {
	logtrace.LogWithFunctionName()
	return sha256.Sum256([]byte(c.TokenSecret))
}
