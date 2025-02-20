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
package verify

import (
	logtrace "ztna-core/ztna/logtrace"
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"ztna-core/ztna/internal"
	"ztna-core/ztna/ztna/cmd/edge"

	"github.com/michaelquigley/pfxlog"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"ztna-core/edge-api/rest_management_api_client"
	"ztna-core/edge-api/rest_management_api_client/identity"
	"ztna-core/edge-api/rest_management_api_client/service"
	"ztna-core/edge-api/rest_management_api_client/service_policy"
	"ztna-core/edge-api/rest_management_api_client/terminator"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/internal/rest/mgmt"

	"ztna-core/sdk-golang/ziti"
	"ztna-core/sdk-golang/ziti/enroll"
)

type traffic struct {
	loginOpts            edge.LoginOptions
	prefix               string
	mode                 string
	cleanup              bool
	verbose              bool
	allowMultipleServers bool

	client       *rest_management_api_client.ZitiEdgeManagement
	svcName      string
	serverIdName string
	clientIdName string
	bindSPName   string
	dialSPName   string
	haEnabled    bool
}

func NewVerifyTraffic(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	t := &traffic{}
	cmd := &cobra.Command{
		Use:   "verify-traffic",
		Short: "Verifies traffic",
		Long:  "A tool to verify traffic can flow over the overlay properly. You must be authenticated to use this tool.",
		Run: func(cmd *cobra.Command, args []string) {
			logLvl := logrus.InfoLevel
			if t.verbose {
				logLvl = logrus.DebugLevel
			}

			pfxlog.GlobalInit(logLvl, pfxlog.DefaultOptions().Color())
			internal.ConfigureLogFormat(logLvl)

			timePrefix := time.Now().Format("2006-01-02-1504")
			if t.prefix == "" {
				if t.mode != "both" {
					log.Warnf("no prefix and mode [%s] is not 'both'. default prefix of %s will be used", t.mode, timePrefix)
				}
				t.prefix = timePrefix
			}
			if t.mode == "" {
				t.mode = "both"
			}

			t.svcName = t.prefix + ".verify-traffic"

			t.serverIdName = t.prefix + ".server"
			extraSeverIdName := ""
			if t.allowMultipleServers {
				extraSeverIdName = fmt.Sprintf("%d", time.Now().UnixNano())
			}
			t.serverIdName = fmt.Sprintf("%s.server%s", t.prefix, extraSeverIdName)
			t.clientIdName = t.prefix + ".client"
			t.bindSPName = t.prefix + ".bind"
			t.dialSPName = t.prefix + ".dial"

			var err error
			t.client, err = t.loginOpts.NewMgmtClient()
			if err != nil {
				log.Fatal(err)
			}

			if t.cleanup {
				log.Info("attempting to cleanup based on parameters. this operation will disconnect the server if it's running.")
				t.cleanupClient()
				t.cleanupServer()
				log.Info("cleanup complete. continuing")
			}

			if t.mode == "both" {
				t.doBoth()
			} else if t.mode == "server" {
				t.doServer(context.Background(), true)
			} else if t.mode == "client" {
				_, c := context.WithCancel(context.Background())
				t.doClient(c)
			} else {
				log.Fatal("no role supplied? should have defaulted to 'both'")
			}
		},
	}

	cmd.Flags().StringVarP(&t.prefix, "prefix", "x", "", "[optional] The prefix to apply to generated objects, necessary when not using the 'both' role.")
	cmd.Flags().StringVarP(&t.mode, "mode", "m", "", "[optional, default 'both'] The mode to perform: server, client, both.")
	cmd.Flags().BoolVar(&t.cleanup, "cleanup", false, "Whether to perform cleanup.")
	cmd.Flags().BoolVar(&t.allowMultipleServers, "allow-multiple-servers", false, "Whether to allows the same server multiple times.")
	cmd.Flags().BoolVar(&t.haEnabled, "ha", false, "Enable high availability mode.")
	cmd.Flags().StringVar(&t.loginOpts.ControllerUrl, "controller-url", "", "The url of the controller")
	_ = cmd.Flags().MarkHidden("ha")

	edge.AddLoginFlags(cmd, &t.loginOpts)
	t.loginOpts.Out = out
	t.loginOpts.Err = errOut

	return cmd
}

func (t *traffic) startServer(ctx context.Context, serviceName string, zitiCfg *ziti.Config) error {
	logtrace.LogWithFunctionName()
	zitiCfg.EnableHa = t.haEnabled
	c, err := ziti.NewContext(zitiCfg)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := c.Listen(serviceName)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("successfully bound service: %s.", serviceName)

	connChan := make(chan net.Conn)
	errChan := make(chan error)
	go func() {
		fmt.Println() // put a line in output for the humans
		log.Info("Server is listening for a connection and will exit when one is received.")
		conn, err := listener.Accept()
		log.Info("Server has accepted a connection and will exit soon.")
		if err != nil {
			errChan <- err
			return
		}
		connChan <- conn
	}()

	select {
	case conn := <-connChan:
		handleConnection(conn)
	case err := <-errChan:
		log.Errorf("Error accepting connection: %v", err)
	case <-ctx.Done():
		log.Info("Server shutting down")
		return ctx.Err()
	}
	_ = listener.Close()
	time.Sleep(1 * time.Second)
	log.Info("Server complete. exiting")
	return nil
}

func handleConnection(conn net.Conn) {
	logtrace.LogWithFunctionName()
	log.Debug("new connection accepted")

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)
	rw := bufio.NewReadWriter(reader, writer)

	line, err := rw.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(line, "verify-traffic test") {
		log.Info("verify-traffic test successfully detected")
	}
	log.Debugf("read : %s", strings.TrimSpace(line))
	resp := fmt.Sprintf("you sent me: %s", line)
	_, _ = rw.WriteString(resp)
	_ = rw.Flush()
	log.Debugf("responding with : %s", strings.TrimSpace(resp))
}

func (t *traffic) startClient(client *rest_management_api_client.ZitiEdgeManagement, serviceName string, zitiCfg *ziti.Config) error {
	logtrace.LogWithFunctionName()
	waitForTerminator(client, serviceName, 10*time.Second)
	zitiCfg.EnableHa = t.haEnabled
	c, err := ziti.NewContext(zitiCfg)
	if err != nil {
		log.Fatal(err)
	}

	foundSvc, ok := c.GetService(serviceName)
	if !ok {
		log.Fatal("error when retrieving all the services for the provided config")
	}
	log.Infof("found service named: %s", *foundSvc.Name)

	svc, err := c.Dial(serviceName) //dial the service using the given name
	if err != nil {
		log.Fatalf("error when dialing service name %s. %v", serviceName, err)
	}
	log.Infof("successfully dialed service: %s.", serviceName)

	zitiReader := bufio.NewReader(svc)
	zitiWriter := bufio.NewWriter(svc)

	text := "verify-traffic test\n"
	bytesRead, err := zitiWriter.WriteString(text)
	_ = zitiWriter.Flush()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Debugf("wrote %d bytes", bytesRead)
	}
	log.Debugf("sent : %s", text)
	read, err := zitiReader.ReadString('\n')
	if err != nil {
		log.Errorf("error reading from reader: %v", err)
	} else {
		log.Debugf("Received: %s", strings.TrimSpace(read))
	}
	return nil
}

func terminatorExists(client *rest_management_api_client.ZitiEdgeManagement, serviceName string) bool {
	logtrace.LogWithFunctionName()
	filter := "service.name=\"" + serviceName + "\""
	params := &terminator.ListTerminatorsParams{
		Filter:  &filter,
		Context: context.Background(),
	}

	resp, err := client.Terminator.ListTerminators(params, nil)
	if err != nil {
		log.Fatal(err)
	}

	return len(resp.Payload.Data) > 0
}

func waitForTerminator(client *rest_management_api_client.ZitiEdgeManagement, serviceName string, timeout time.Duration) bool {
	logtrace.LogWithFunctionName()
	log.Infof("waiting %s for terminator for service: %s", timeout, serviceName)
	startTime := time.Now()
	for {
		if terminatorExists(client, serviceName) {
			log.Infof("found terminator for service: %s", serviceName)
			return true
		}
		if time.Since(startTime) >= timeout {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Fatalf("terminator not found for service: %s", serviceName)
	return false
}

func createIdentity(client *rest_management_api_client.ZitiEdgeManagement, name string, roleAttributes rest_model.Attributes) *identity.CreateIdentityCreated {
	logtrace.LogWithFunctionName()
	falseVar := false
	usrType := rest_model.IdentityTypeUser
	i := &rest_model.IdentityCreate{
		Enrollment: &rest_model.IdentityCreateEnrollment{
			Ott: true,
		},
		IsAdmin:        &falseVar,
		Name:           &name,
		RoleAttributes: &roleAttributes,
		Type:           &usrType,
	}
	p := identity.NewCreateIdentityParams()
	p.Identity = i

	// Create the identity
	ident, err := client.Identity.CreateIdentity(p, nil)
	if err != nil {
		id := mgmt.IdentityFromFilter(client, mgmt.NameFilter(name))
		if id != nil {
			log.Fatalf("Identity named %s exists. Remove the identity before trying again or use --cleanup.", name)
		} else {
			log.Fatalf("Failed to create the identity: %v", err)
		}
	}
	return ident
}

func createServicePolicy(client *rest_management_api_client.ZitiEdgeManagement, name string, servType rest_model.DialBind, identityRoles rest_model.Roles, serviceRoles rest_model.Roles) *rest_model.CreateLocation {
	logtrace.LogWithFunctionName()
	defaultSemantic := rest_model.SemanticAllOf
	servicePolicy := &rest_model.ServicePolicyCreate{
		IdentityRoles: identityRoles,
		Name:          &name,
		Semantic:      &defaultSemantic,
		ServiceRoles:  serviceRoles,
		Type:          &servType,
	}
	params := &service_policy.CreateServicePolicyParams{
		Policy:  servicePolicy,
		Context: context.Background(),
	}
	params.SetTimeout(5 * time.Second)
	resp, err := client.ServicePolicy.CreateServicePolicy(params, nil)
	if resp == nil || err != nil {
		log.Fatalf("Failed to create service policy: %s", name)
		return nil
	}
	return resp.Payload.Data
}

func createService(client *rest_management_api_client.ZitiEdgeManagement, name string, serviceConfigs []string, roles rest_model.Attributes) *rest_model.CreateLocation {
	logtrace.LogWithFunctionName()
	encryptOn := true
	serviceCreate := &rest_model.ServiceCreate{
		Configs:            serviceConfigs,
		EncryptionRequired: &encryptOn,
		MaxIdleTimeMillis:  0,
		Name:               &name,
		RoleAttributes:     roles,
		Tags:               nil,
		TerminatorStrategy: "",
	}
	serviceParams := &service.CreateServiceParams{
		Service: serviceCreate,
		Context: context.Background(),
	}
	serviceParams.SetTimeout(5 * time.Second)
	resp, err := client.Service.CreateService(serviceParams, nil)
	if resp == nil || err != nil {
		log.Fatalf("Failed to create service: %s. %v", name, err)
		return nil
	}
	return resp.Payload.Data
}

func deleteIdentity(client *rest_management_api_client.ZitiEdgeManagement, toDelete *rest_model.IdentityDetail) {
	logtrace.LogWithFunctionName()
	if toDelete == nil {
		return
	}
	idToDel := *toDelete.ID
	deleteParams := &identity.DeleteIdentityParams{
		ID: idToDel,
	}
	deleteParams.SetTimeout(5 * time.Second)
	_, err := client.Identity.DeleteIdentity(deleteParams, nil)
	if err != nil {
		log.Errorf("Failed to delete identity: %s. %v", idToDel, err)
	}
}

func deleteService(client *rest_management_api_client.ZitiEdgeManagement, toDelete *rest_model.ServiceDetail) {
	logtrace.LogWithFunctionName()
	if toDelete == nil {
		return
	}
	idToDel := *toDelete.ID
	deleteParams := &service.DeleteServiceParams{
		ID: idToDel,
	}
	deleteParams.SetTimeout(5 * time.Second)
	_, err := client.Service.DeleteService(deleteParams, nil)
	if err != nil {
		log.Errorf("Failed to delete service: %s. %v", idToDel, err)
	}
}

func deleteServicePolicy(client *rest_management_api_client.ZitiEdgeManagement, sp *rest_model.ServicePolicyDetail) {
	logtrace.LogWithFunctionName()
	if sp == nil {
		return
	}
	id := *sp.ID
	deleteParams := &service_policy.DeleteServicePolicyParams{
		ID: id,
	}
	deleteParams.SetTimeout(5 * time.Second)
	_, err := client.ServicePolicy.DeleteServicePolicy(deleteParams, nil)
	if err != nil {
		log.Errorf("Failed to delete the service policy: %s. %v", id, err)
	}
}

func enrollIdentity(client *rest_management_api_client.ZitiEdgeManagement, id string) *ziti.Config {
	logtrace.LogWithFunctionName()
	// Get the identity object
	params := &identity.DetailIdentityParams{
		Context: context.Background(),
		ID:      id,
	}
	params.SetTimeout(5 * time.Second)
	resp, err := client.Identity.DetailIdentity(params, nil)

	if err != nil {
		log.Fatal(err)
	}

	// Enroll the identity
	tkn, _, err := enroll.ParseToken(resp.Payload.Data.Enrollment.Ott.JWT)
	if err != nil {
		log.Fatal(err)
	}

	flags := enroll.EnrollmentFlags{
		Token:  tkn,
		KeyAlg: "EC",
	}
	conf, err := enroll.Enroll(flags)

	if err != nil {
		log.Fatal(err)
	}

	return conf
}

func (t *traffic) bindAttr() string {
	logtrace.LogWithFunctionName()
	return t.svcName + ".binders"
}

func (t *traffic) dialAttr() string {
	logtrace.LogWithFunctionName()
	return t.svcName + ".dialers"
}

func (t *traffic) svcAttr() string {
	logtrace.LogWithFunctionName()
	return t.svcName
}

func (t *traffic) configureService() {
	logtrace.LogWithFunctionName()
	svc := mgmt.ServiceFromFilter(t.client, mgmt.NameFilter(t.svcName))
	if svc != nil && t.allowMultipleServers {
		log.Debugf("service already exists. not creating: %s", t.svcName)
	} else {
		_ = createService(t.client, t.svcName, nil, []string{t.svcAttr()})
	}

	bind := mgmt.ServicePolicyFromFilter(t.client, mgmt.NameFilter(t.bindSPName))
	if bind != nil && t.allowMultipleServers {
		log.Debugf("service policy already exists. not creating: %s", t.bindSPName)
	} else {
		_ = createServicePolicy(t.client, t.bindSPName, rest_model.DialBindBind, rest_model.Roles{"#" + t.bindAttr()}, rest_model.Roles{"#" + t.svcAttr()})
	}

	dial := mgmt.ServicePolicyFromFilter(t.client, mgmt.NameFilter(t.dialSPName))
	if dial != nil && t.allowMultipleServers {
		log.Debugf("service policy already exists. not creating: %s", t.dialSPName)
	} else {
		_ = createServicePolicy(t.client, t.dialSPName, rest_model.DialBindDial, rest_model.Roles{"#" + t.dialAttr()}, rest_model.Roles{"#" + t.svcAttr()})
	}
}

func (t *traffic) configureServer() *ziti.Config {
	logtrace.LogWithFunctionName()
	serverIdent := createIdentity(t.client, t.serverIdName, []string{t.bindAttr()})
	return enrollIdentity(t.client, serverIdent.Payload.Data.ID)
}

func (t *traffic) configureClient() *ziti.Config {
	logtrace.LogWithFunctionName()
	clientIdent := createIdentity(t.client, t.clientIdName, []string{t.dialAttr()})
	return enrollIdentity(t.client, clientIdent.Payload.Data.ID)
}

func (t *traffic) cleanupServer() {
	logtrace.LogWithFunctionName()
	if t.allowMultipleServers {
		if terminatorExists(t.client, t.svcName) {
			log.Debugf("found terminator for service: %s. cleanup will be skipped.", t.svcName)
			return
		}
	}
	dial := mgmt.ServicePolicyFromFilter(t.client, mgmt.NameFilter(t.dialSPName))
	bind := mgmt.ServicePolicyFromFilter(t.client, mgmt.NameFilter(t.bindSPName))
	deleteServicePolicy(t.client, dial)
	deleteServicePolicy(t.client, bind)
	svc := mgmt.ServiceFromFilter(t.client, mgmt.NameFilter(t.svcName))
	deleteService(t.client, svc)

	id := mgmt.IdentityFromFilter(t.client, mgmt.NameFilter(t.serverIdName))
	deleteIdentity(t.client, id)
}

func (t *traffic) cleanupClient() {
	logtrace.LogWithFunctionName()
	id := mgmt.IdentityFromFilter(t.client, mgmt.NameFilter(t.clientIdName))
	deleteIdentity(t.client, id)
}

func (t *traffic) doBoth() {
	logtrace.LogWithFunctionName()
	t.configureService()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer wg.Done()
		t.doServer(ctx, false)
	}()
	go func() {
		defer wg.Done()
		t.doClient(cancel)
	}()
	wg.Wait()
}

func (t *traffic) doServer(ctx context.Context, configureServices bool) {
	logtrace.LogWithFunctionName()
	if configureServices {
		t.configureService()
	}
	serverCfg := t.configureServer()
	defer t.cleanupServer()
	if err := t.startServer(ctx, t.svcName, serverCfg); err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
}

func (t *traffic) doClient(cancel context.CancelFunc) {
	logtrace.LogWithFunctionName()
	clientCfg := t.configureClient()
	defer t.cleanupClient()
	if err := t.startClient(t.client, t.svcName, clientCfg); err != nil {
		log.Fatal(err)
	}

	log.Debug("client received expected response. stopping server if it's running")
	cancel() //end the server
	time.Sleep(1 * time.Second)
	log.Info("client complete")
}
