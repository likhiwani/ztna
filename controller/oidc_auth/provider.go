package oidc_auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"ztna-core/ztna/common"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/logtrace"

	"github.com/gorilla/mux"
	"github.com/michaelquigley/pfxlog"
	"github.com/pkg/errors"
	"github.com/zitadel/oidc/v2/pkg/op"
	"golang.org/x/text/language"
)

const (
	pathLoggedOut              = "/oidc/logged-out"
	WellKnownOidcConfiguration = "/.well-known/openid-configuration"

	SourceTypeOidc = "oidc_auth"

	AuthMethodPassword = model.AuthMethodPassword
	AuthMethodExtJwt   = model.AuthMethodExtJwt
	AuthMethodCert     = db.MethodAuthenticatorCert

	AuthMethodSecondaryTotp   = "totp"
	AuthMethodSecondaryExtJwt = "ejs"
)

// NewNativeOnlyOP creates an OIDC Provider that allows native clients and only the AutCode PKCE flow.
func NewNativeOnlyOP(ctx context.Context, env model.Env, config Config) (http.Handler, error) {
	logtrace.LogWithFunctionName()
	cert, kid, method := env.GetServerCert()
	config.Storage = NewStorage(kid, cert.Leaf.PublicKey, cert.PrivateKey, method, &config, env)

	handlers := map[string]http.Handler{}

	for _, issuer := range config.Issuers {
		issuerUrl := "https://" + issuer + "/oidc"
		provider, err := newOidcProvider(ctx, issuerUrl, config)
		if err != nil {
			return nil, fmt.Errorf("could not create OpenIdProvider: %w", err)
		}

		oidcHandler, err := newHttpRouter(provider, config)

		openzitiClient := NativeClient(common.ClaimClientIdOpenZiti, config.RedirectURIs, config.PostLogoutURIs)
		openzitiClient.idTokenDuration = config.IdTokenDuration
		openzitiClient.loginURL = newLoginResolver(config.Storage)
		config.Storage.AddClient(openzitiClient)

		//backwards compatibility client w/ early HA SDKs. Should be removed by the time HA is GA'ed.
		nativeClient := NativeClient(common.ClaimLegacyNative, config.RedirectURIs, config.PostLogoutURIs)
		nativeClient.idTokenDuration = config.IdTokenDuration
		nativeClient.loginURL = newLoginResolver(config.Storage)
		config.Storage.AddClient(nativeClient)

		if err != nil {
			return nil, err
		}

		handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			r := request.WithContext(context.WithValue(request.Context(), contextKeyHttpRequest, request))
			r = request.WithContext(context.WithValue(r.Context(), contextKeyTokenState, &TokenState{}))
			r = request.WithContext(op.ContextWithIssuer(r.Context(), issuerUrl))

			oidcHandler.ServeHTTP(writer, r)
		})

		hostsToHandle := getHandledHostnames(issuer)

		for _, hostToHandle := range hostsToHandle {
			handlers[hostToHandle] = handler
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler, ok := handlers[r.Host]

		if !ok {
			http.NotFound(w, r)
			return
		}

		handler.ServeHTTP(w, r)
	}), nil

}

func getHandledHostnames(issuer string) []string {
	logtrace.LogWithFunctionName()
	const (
		DefaultTlsPort = "443"
		LocalhostName  = "localhost"
		LocalhostIpv4  = "127.0.0.1"
		LocalhostIpv6  = "::1"
	)
	hostsToHandle := map[string]struct{}{
		issuer: {},
	}

	hostWithoutPort, port, err := net.SplitHostPort(issuer)
	if err != nil {
		var ret []string
		for host := range hostsToHandle {
			ret = append(ret, host)
		}

		return ret
	}

	shouldHandleDefaultPort := port == DefaultTlsPort
	if shouldHandleDefaultPort {

		ip := net.ParseIP(hostWithoutPort)
		isIpv6 := ip != nil && ip.To4() == nil

		if isIpv6 {
			//ipv6 in urls always requires brackets even w/ default ports
			hostsToHandle["["+hostWithoutPort+"]"] = struct{}{}
		} else {
			hostsToHandle[hostWithoutPort] = struct{}{}
		}

	}

	//local address in use, translate as needed
	if hostWithoutPort == LocalhostName || hostWithoutPort == LocalhostIpv4 || hostWithoutPort == "::1" {
		hostsToHandle[net.JoinHostPort(LocalhostName, port)] = struct{}{}
		hostsToHandle[net.JoinHostPort(LocalhostIpv4, port)] = struct{}{}
		hostsToHandle[net.JoinHostPort(LocalhostIpv6, port)] = struct{}{}

		if shouldHandleDefaultPort {
			hostsToHandle[LocalhostName] = struct{}{}
			hostsToHandle[LocalhostIpv4] = struct{}{}

			//ipv6 in urls always requires brackets even w/ default ports
			hostsToHandle["["+LocalhostIpv6+"]"] = struct{}{}

		}
	}

	var ret []string
	for host := range hostsToHandle {
		ret = append(ret, host)
	}

	return ret
}

// newHttpRouter creates an OIDC HTTP router
func newHttpRouter(provider op.OpenIDProvider, config Config) (*mux.Router, error) {
	logtrace.LogWithFunctionName()
	if config.TokenSecret == "" {
		return nil, errors.New("token secret must not be empty")
	}

	router := mux.NewRouter()

	router.HandleFunc(pathLoggedOut, func(w http.ResponseWriter, req *http.Request) {
		_, err := w.Write([]byte("signed out successfully"))
		if err != nil {
			pfxlog.Logger().Errorf("error serving logged out page: %v", err)
		}
	})

	loginRouter := newLogin(config.Storage, op.AuthCallbackURL(provider), op.NewIssuerInterceptor(provider.IssuerFromRequest))

	router.Handle("/oidc/"+WellKnownOidcConfiguration, http.StripPrefix("/oidc", provider.HttpHandler()))
	router.Handle(WellKnownOidcConfiguration, provider.HttpHandler())

	router.PathPrefix("/oidc/login").Handler(http.StripPrefix("/oidc/login", loginRouter.router))

	router.PathPrefix("/oidc").Handler(http.StripPrefix("/oidc", provider.HttpHandler()))

	return router, nil
}

// newOidcProvider will create an OpenID Provider that allows refresh tokens, authentication via form post and basic auth, and support request object params
func newOidcProvider(_ context.Context, issuer string, oidcConfig Config) (op.OpenIDProvider, error) {
	logtrace.LogWithFunctionName()
	config := &op.Config{
		CryptoKey:                oidcConfig.Secret(),
		DefaultLogoutRedirectURI: pathLoggedOut,
		CodeMethodS256:           true,
		AuthMethodPost:           true,
		AuthMethodPrivateKeyJWT:  true,
		GrantTypeRefreshToken:    true,
		RequestObjectSupported:   true,
		SupportedUILocales:       []language.Tag{language.English},
	}

	handler, err := op.NewOpenIDProvider(issuer, config, oidcConfig.Storage)

	if err != nil {
		return nil, err
	}
	return handler, nil
}

// newLoginResolver returns func capable of determining default login URLs based on authId
func newLoginResolver(storage Storage) func(string) string {
	logtrace.LogWithFunctionName()
	return func(authId string) string {
		authRequest, err := storage.GetAuthRequest(authId)

		if err != nil || authRequest == nil {
			return passwordLoginUrl + authId
		}

		switch authRequest.RequestedMethod {
		case AuthMethodPassword:
			return passwordLoginUrl + authId
		case AuthMethodExtJwt:
			return extJwtLoginUrl + authId
		case AuthMethodCert:
			return certLoginUrl + authId
		}

		if len(authRequest.PeerCerts) > 0 {
			return certLoginUrl + authId
		}

		return passwordLoginUrl + authId
	}
}
