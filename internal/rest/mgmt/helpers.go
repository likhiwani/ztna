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

package mgmt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
	logtrace "ztna-core/ztna/logtrace"

	"ztna-core/edge-api/rest_management_api_client"
	"ztna-core/edge-api/rest_management_api_client/auth_policy"
	"ztna-core/edge-api/rest_management_api_client/certificate_authority"
	"ztna-core/edge-api/rest_management_api_client/config"
	rest_mgmt "ztna-core/edge-api/rest_management_api_client/current_api_session"
	"ztna-core/edge-api/rest_management_api_client/edge_router"
	"ztna-core/edge-api/rest_management_api_client/edge_router_policy"
	"ztna-core/edge-api/rest_management_api_client/external_jwt_signer"
	"ztna-core/edge-api/rest_management_api_client/identity"
	"ztna-core/edge-api/rest_management_api_client/posture_checks"
	"ztna-core/edge-api/rest_management_api_client/service"
	"ztna-core/edge-api/rest_management_api_client/service_edge_router_policy"
	"ztna-core/edge-api/rest_management_api_client/service_policy"
	"ztna-core/edge-api/rest_model"
	"ztna-core/edge-api/rest_util"
	"ztna-core/ztna/ztna/util"

	log "github.com/sirupsen/logrus"
)

const (
	DefaultTimeout = 5 * time.Second
)

func IdentityFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.IdentityDetail {
	logtrace.LogWithFunctionName()
	params := &identity.ListIdentitiesParams{
		Filter:  &filter,
		Context: context.Background(),
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.Identity.ListIdentities(params, nil)
	if err != nil {
		log.Debugf("Could not obtain an ID for the identity with filter %s: %v", filter, err)
		return nil
	}

	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func ServiceFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.ServiceDetail {
	logtrace.LogWithFunctionName()
	params := &service.ListServicesParams{
		Filter:  &filter,
		Context: context.Background(),
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.Service.ListServices(params, nil)
	if err != nil {
		log.Debugf("Could not obtain an ID for the service with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func ServicePolicyFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.ServicePolicyDetail {
	logtrace.LogWithFunctionName()
	params := &service_policy.ListServicePoliciesParams{
		Filter:  &filter,
		Context: context.Background(),
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.ServicePolicy.ListServicePolicies(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the service policy with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func AuthPolicyFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.AuthPolicyDetail {
	logtrace.LogWithFunctionName()
	params := &auth_policy.ListAuthPoliciesParams{
		Filter:  &filter,
		Context: context.Background(),
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.AuthPolicy.ListAuthPolicies(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the auth policy with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func CertificateAuthorityFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.CaDetail {
	logtrace.LogWithFunctionName()
	params := &certificate_authority.ListCasParams{
		Filter:  &filter,
		Context: context.Background(),
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.CertificateAuthority.ListCas(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the certificate authority with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func ConfigTypeFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.ConfigTypeDetail {
	logtrace.LogWithFunctionName()
	params := &config.ListConfigTypesParams{
		Filter:  &filter,
		Context: context.Background(),
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.Config.ListConfigTypes(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the config type with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func ConfigFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.ConfigDetail {
	logtrace.LogWithFunctionName()
	params := &config.ListConfigsParams{
		Filter:  &filter,
		Context: context.Background(),
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.Config.ListConfigs(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the config with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func ExternalJWTSignerFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.ExternalJWTSignerDetail {
	logtrace.LogWithFunctionName()
	params := &external_jwt_signer.ListExternalJWTSignersParams{
		Filter:  &filter,
		Context: context.Background(),
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.ExternalJWTSigner.ListExternalJWTSigners(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the external jwt signer with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func PostureCheckFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.PostureCheckDetail {
	logtrace.LogWithFunctionName()
	params := &posture_checks.ListPostureChecksParams{
		Filter:  &filter,
		Context: context.Background(),
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.PostureChecks.ListPostureChecks(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the posture check with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || len(resp.Payload.Data()) == 0 {
		return nil
	}
	return &resp.Payload.Data()[0]
}

func EdgeRouterPolicyFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.EdgeRouterPolicyDetail {
	logtrace.LogWithFunctionName()
	params := &edge_router_policy.ListEdgeRouterPoliciesParams{
		Filter: &filter,
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.EdgeRouterPolicy.ListEdgeRouterPolicies(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the edge router policies with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func EdgeRouterFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.EdgeRouterDetail {
	logtrace.LogWithFunctionName()
	params := &edge_router.ListEdgeRoutersParams{
		Filter: &filter,
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.EdgeRouter.ListEdgeRouters(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the edge routers with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func ServiceEdgeRouterPolicyFromFilter(client *rest_management_api_client.ZitiEdgeManagement, filter string) *rest_model.ServiceEdgeRouterPolicyDetail {
	logtrace.LogWithFunctionName()
	params := &service_edge_router_policy.ListServiceEdgeRouterPoliciesParams{
		Filter: &filter,
	}
	params.SetTimeout(DefaultTimeout)
	resp, err := client.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(params, nil)
	if err != nil {
		log.Errorf("Could not obtain an ID for the ServiceEdgeRouterPolicy routers with filter %s: %v", filter, err)
		return nil
	}
	if resp == nil || resp.Payload == nil || resp.Payload.Data == nil || len(resp.Payload.Data) == 0 {
		return nil
	}
	return resp.Payload.Data[0]
}

func NameFilter(name string) string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("name = \"%s\"", name)
}

func NewClient() (*rest_management_api_client.ZitiEdgeManagement, error) {
	logtrace.LogWithFunctionName()
	cachedCreds, _, loadErr := util.LoadRestClientConfig()
	if loadErr != nil {
		return nil, loadErr
	}

	cachedId := cachedCreds.EdgeIdentities[cachedCreds.Default] //only support default for now
	if cachedId == nil {
		return nil, errors.New("no identity found")
	}

	caPool := x509.NewCertPool()
	if _, cacertErr := os.Stat(cachedId.CaCert); cacertErr == nil {
		rootPemData, err := os.ReadFile(cachedId.CaCert)
		if err != nil {
			return nil, err
		}
		caPool.AppendCertsFromPEM(rootPemData)
	} else {
		return nil, errors.New("CA cert file not found in config file")
	}

	tlsConfig := &tls.Config{
		RootCAs: caPool,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Assign the transport to the default HTTP client
	http.DefaultClient = &http.Client{
		Transport: transport,
	}
	c, e := rest_util.NewEdgeManagementClientWithToken(http.DefaultClient, cachedId.Url, cachedId.Token)
	if e != nil {
		return nil, e
	}

	apiSessionParams := &rest_mgmt.GetCurrentAPISessionParams{
		Context: context.Background(),
	}
	_, authErr := c.CurrentAPISession.GetCurrentAPISession(apiSessionParams, nil)
	if authErr != nil {
		return nil, errors.New("client not authenticated. login with 'ziti edge login' before executing")
	}
	return c, nil
}
