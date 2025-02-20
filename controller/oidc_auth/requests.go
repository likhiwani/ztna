package oidc_auth

import (
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"time"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/common"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/stringz"

	"github.com/zitadel/oidc/v2/pkg/oidc"
)

// AuthRequest represents an OIDC authentication request and implements op.AuthRequest
type AuthRequest struct {
	oidc.AuthRequest
	Id                    string
	CreationDate          time.Time
	IdentityId            string
	AuthTime              time.Time
	ApiSessionId          string
	SecondaryTotpRequired bool
	SecondaryExtJwtSigner *model.ExternalJwtSigner
	ConfigTypes           []string
	Amr                   map[string]struct{}

	PeerCerts           []*x509.Certificate
	RequestedMethod     string
	BearerTokenDetected bool
	SdkInfo             *rest_model.SdkInfo
	EnvInfo             *rest_model.EnvInfo
	RemoteAddress       string
	IsCertExtendable    bool
}

// GetID returns an AuthRequest's ID and implements op.AuthRequest
func (a *AuthRequest) GetID() string {
	logtrace.LogWithFunctionName()
	return a.Id
}

// GetACR returns the authentication class reference provided by client and implements oidc.AuthRequest
// All ACRs are currently ignored.
func (a *AuthRequest) GetACR() string {
	logtrace.LogWithFunctionName()
	return ""
}

// GetAMR returns the authentication method references the authentication has undergone and implements op.AuthRequest
func (a *AuthRequest) GetAMR() []string {
	logtrace.LogWithFunctionName()
	result := make([]string, len(a.Amr))
	i := 0
	for k := range a.Amr {
		result[i] = k
		i = i + 1
	}
	return result
}

// HasFullAuth returns true if an authentication request has passed all primary and secondary authentications.
func (a *AuthRequest) HasFullAuth() bool {
	logtrace.LogWithFunctionName()
	return a.HasPrimaryAuth() && a.HasSecondaryAuth()
}

// HasPrimaryAuth returns true if a primary authentication mechanism has been passed.
func (a *AuthRequest) HasPrimaryAuth() bool {
	logtrace.LogWithFunctionName()
	return a.HasAmr(AuthMethodCert) || a.HasAmr(AuthMethodPassword) || a.HasAmr(AuthMethodExtJwt)
}

// HasSecondaryAuth returns true if all applicable secondary authentications have been passed
func (a *AuthRequest) HasSecondaryAuth() bool {
	logtrace.LogWithFunctionName()
	return (!a.SecondaryTotpRequired || a.HasAmr(AuthMethodSecondaryTotp)) &&
		(a.SecondaryExtJwtSigner == nil || a.HasAmrExtJwtId(a.SecondaryExtJwtSigner.Id))
}

// HasAmr returns true if the supplied amr is present
func (a *AuthRequest) HasAmr(amr string) bool {
	logtrace.LogWithFunctionName()
	_, found := a.Amr[amr]
	return found
}

func (a *AuthRequest) HasAmrExtJwtId(id string) bool {
	logtrace.LogWithFunctionName()
	return a.HasAmr(AuthMethodSecondaryExtJwt + ":" + id)
}

// AddAmr adds the supplied amr
func (a *AuthRequest) AddAmr(amr string) {
	logtrace.LogWithFunctionName()
	if a.Amr == nil {
		a.Amr = map[string]struct{}{}
	}
	a.Amr[amr] = struct{}{}
}

// GetAudience returns all current audience targets and implements op.AuthRequest
func (a *AuthRequest) GetAudience() []string {
	logtrace.LogWithFunctionName()
	return []string{a.ClientID}
}

// GetAuthTime returns the time at which authentication has occurred and implements op.AuthRequest
func (a *AuthRequest) GetAuthTime() time.Time {
	logtrace.LogWithFunctionName()
	return a.AuthTime
}

// GetClientID returns the client id requested and implements op.AuthRequest
func (a *AuthRequest) GetClientID() string {
	logtrace.LogWithFunctionName()
	return a.ClientID
}

// GetCodeChallenge returns the rp supplied code change and implements op.AuthRequest
func (a *AuthRequest) GetCodeChallenge() *oidc.CodeChallenge {
	logtrace.LogWithFunctionName()
	return &oidc.CodeChallenge{
		Challenge: a.CodeChallenge,
		Method:    a.CodeChallengeMethod,
	}
}

// GetNonce returns the rp supplied nonce and implements op.AuthRequest
func (a *AuthRequest) GetNonce() string {
	logtrace.LogWithFunctionName()
	return a.Nonce
}

// GetRedirectURI returns the rp supplied redirect target and implements op.AuthRequest
func (a *AuthRequest) GetRedirectURI() string {
	logtrace.LogWithFunctionName()
	return a.RedirectURI
}

// GetResponseType returns the rp supplied response type and implements op.AuthRequest
func (a *AuthRequest) GetResponseType() oidc.ResponseType {
	logtrace.LogWithFunctionName()
	return a.ResponseType
}

// GetResponseMode is not supported and all tokens are turned via query string and implements op.AuthRequest
func (a *AuthRequest) GetResponseMode() oidc.ResponseMode {
	logtrace.LogWithFunctionName()
	return ""
}

// GetScopes returns the current scopes and implements op.AuthRequest
func (a *AuthRequest) GetScopes() []string {
	logtrace.LogWithFunctionName()
	return a.Scopes
}

// GetState returns the rp provided state and implements op.AuthRequest
func (a *AuthRequest) GetState() string {
	logtrace.LogWithFunctionName()
	return a.State
}

// GetSubject returns the target subject and implements op.AuthRequest
func (a *AuthRequest) GetSubject() string {
	logtrace.LogWithFunctionName()
	return a.IdentityId
}

// Done returns true once authentication has been completed and implements op.AuthRequest
func (a *AuthRequest) Done() bool {
	logtrace.LogWithFunctionName()
	return a.HasFullAuth()
}

func (a *AuthRequest) GetCertFingerprints() []string {
	logtrace.LogWithFunctionName()
	var prints []string

	for _, cert := range a.PeerCerts {
		prints = append(prints, fmt.Sprintf("%x", sha1.Sum(cert.Raw)))
	}

	return prints
}

func (a *AuthRequest) NeedsTotp() bool {
	logtrace.LogWithFunctionName()
	return a.SecondaryTotpRequired && !a.HasAmr(AuthMethodSecondaryTotp)
}

func (a *AuthRequest) NeedsSecondaryExtJwt() bool {
	logtrace.LogWithFunctionName()
	return a.SecondaryExtJwtSigner != nil && !a.HasAmrExtJwtId(a.SecondaryExtJwtSigner.Id)
}

func (a *AuthRequest) GetAuthQueries() []*rest_model.AuthQueryDetail {
	logtrace.LogWithFunctionName()
	var authQueries []*rest_model.AuthQueryDetail

	if a.NeedsTotp() {
		provider := rest_model.MfaProvidersZiti
		authQueries = append(authQueries, &rest_model.AuthQueryDetail{
			Format:     rest_model.MfaFormatsNumeric,
			HTTPMethod: "POST",
			HTTPURL:    "./oidc/login/totp",
			MaxLength:  8,
			MinLength:  6,
			Provider:   &provider,
			TypeID:     rest_model.AuthQueryTypeTOTP,
		})
	}

	if a.NeedsSecondaryExtJwt() {
		provider := rest_model.MfaProvidersURL
		authQueries = append(authQueries, &rest_model.AuthQueryDetail{
			ClientID: stringz.OrEmpty(a.SecondaryExtJwtSigner.ClientId),
			HTTPURL:  stringz.OrEmpty(a.SecondaryExtJwtSigner.ExternalAuthUrl),
			Scopes:   a.SecondaryExtJwtSigner.Scopes,
			Provider: &provider,
			ID:       a.SecondaryExtJwtSigner.Id,
			TypeID:   rest_model.AuthQueryTypeEXTDashJWT,
		})
	}

	return authQueries
}

// RefreshTokenRequest is a wrapper around RefreshClaims to avoid collisions between go-jwt interface requirements and
// zitadel oidc interface names. Implements zitadel op.RefreshTokenRequest
type RefreshTokenRequest struct {
	common.RefreshClaims
}

// GetAMR implements op.RefreshTokenRequest
func (r *RefreshTokenRequest) GetAMR() []string {
	logtrace.LogWithFunctionName()
	return r.AuthenticationMethodsReferences
}

// GetAudience implements op.RefreshTokenRequest
func (r *RefreshTokenRequest) GetAudience() []string {
	logtrace.LogWithFunctionName()
	return r.Audience
}

// GetAuthTime implements op.RefreshTokenRequest
func (r *RefreshTokenRequest) GetAuthTime() time.Time {
	logtrace.LogWithFunctionName()
	return r.AuthTime.AsTime()
}

// GetClientID implements op.RefreshTokenRequest
func (r *RefreshTokenRequest) GetClientID() string {
	logtrace.LogWithFunctionName()
	return r.ClientID
}

// GetScopes implements op.RefreshTokenRequest
func (r *RefreshTokenRequest) GetScopes() []string {
	logtrace.LogWithFunctionName()
	return r.Scopes
}

// GetSubject implements op.RefreshTokenRequest
func (r *RefreshTokenRequest) GetSubject() string {
	logtrace.LogWithFunctionName()
	return r.Subject
}

// SetCurrentScopes implements op.RefreshTokenRequest
func (r *RefreshTokenRequest) SetCurrentScopes(scopes []string) {
	logtrace.LogWithFunctionName()
	r.Scopes = scopes
}

func (r *RefreshTokenRequest) GetCertFingerprints() []string {
	logtrace.LogWithFunctionName()
	return r.CertFingerprints
}
