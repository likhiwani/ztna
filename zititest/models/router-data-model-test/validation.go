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

package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"time"

	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/common/pb/mgmt_pb"
	logtrace "ztna-core/ztna/logtrace"
	"ztna-core/ztna/zitirest"
	"ztna-core/ztna/zititest/zitilab/chaos"
	"ztna-core/ztna/zititest/zitilab/models"
	"ztna-core/ztna/ztna/util"

	"github.com/go-openapi/runtime"
	"github.com/google/uuid"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/channel/v3/protobufs"
	"github.com/openziti/fablab/kernel/lib/parallel"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/foundation/v2/errorz"
	ptrutil "github.com/openziti/foundation/v2/util"
	"google.golang.org/protobuf/proto"
)

type CtrlClients struct {
	ctrls   []*zitirest.Clients
	ctrlMap map[string]*zitirest.Clients
	sync.Mutex
}

func (self *CtrlClients) init(run model.Run, selector string) error {
	logtrace.LogWithFunctionName()
	self.ctrlMap = map[string]*zitirest.Clients{}
	ctrls := run.GetModel().SelectComponents(selector)
	resultC := make(chan struct {
		err     error
		id      string
		clients *zitirest.Clients
	}, len(ctrls))

	for _, ctrl := range ctrls {
		go func() {
			clients, err := chaos.EnsureLoggedIntoCtrl(run, ctrl, time.Minute)
			resultC <- struct {
				err     error
				id      string
				clients *zitirest.Clients
			}{
				err:     err,
				id:      ctrl.Id,
				clients: clients,
			}
		}()
	}

	for i := 0; i < len(ctrls); i++ {
		result := <-resultC
		if result.err != nil {
			return result.err
		}
		self.ctrls = append(self.ctrls, result.clients)
		self.ctrlMap[result.id] = result.clients
	}
	return nil
}

func (self *CtrlClients) getRandomCtrl() *zitirest.Clients {
	logtrace.LogWithFunctionName()
	return self.ctrls[rand.Intn(len(self.ctrls))]
}

func (self *CtrlClients) getCtrl(id string) *zitirest.Clients {
	logtrace.LogWithFunctionName()
	return self.ctrlMap[id]
}

// start with a random scenario then cycle through them
var scenarioCounter = rand.Intn(7)

func sowChaos(run model.Run) error {
	logtrace.LogWithFunctionName()
	ctrls := &CtrlClients{}
	if err := ctrls.init(run, ".ctrl"); err != nil {
		return err
	}

	var err error

	tasks, lastTasks, err := getServiceAndConfigChaosTasks(run, ctrls)
	if err != nil {
		return err
	}

	applyTasks := func(f func(run model.Run, ctrls *CtrlClients) ([]parallel.LabeledTask, error)) {
		var t []parallel.LabeledTask
		if err == nil {
			t, err = f(run, ctrls)
			if err == nil {
				tasks = append(tasks, t...)
			}
		}
	}

	applyTasks(getRestartTasks)
	applyTasks(getIdentityChaosTasks)
	applyTasks(getServicePolicyChaosTasks)
	applyTasks(getPostureTasks)

	if err != nil {
		return err
	}

	chaos.Randomize(tasks)
	tasks = append(tasks, lastTasks...)

	retryPolicy := func(task parallel.LabeledTask, attempt int, err error) parallel.ErrorAction {
		var apiErr util.ApiErrorPayload
		var msg string
		if errors.As(err, &apiErr) {
			if strings.HasPrefix(task.Type(), "delete.") {
				if apiErr.GetPayload().Error.Code == errorz.NotFoundCode {
					return parallel.ErrActionIgnore
				}
			} else if strings.HasPrefix(task.Type(), "create.") && attempt > 1 {
				if apiErr.GetPayload().Error.Code == errorz.CouldNotValidateCode {
					return parallel.ErrActionIgnore
				}
			}
			msg = apiErr.GetPayload().Error.Message
		}

		log := pfxlog.Logger().WithField("attempt", attempt).WithError(err).WithField("task", task.Label())

		var runtimeErr *runtime.APIError
		if errors.As(err, &runtimeErr) {
			if cp, ok := runtimeErr.Response.(runtime.ClientResponse); ok {
				body, _ := io.ReadAll(cp.Body())
				log.WithField("msg", cp.Message()).WithField("body", string(body)).Error("runtime error")
			}
		}

		if attempt > 3 {
			return parallel.ErrActionReport
		}
		if msg != "" {
			log = log.WithField("msg", msg)
		}
		log.Error("action failed, retrying")
		time.Sleep(time.Duration(attempt*10) * time.Second)
		return parallel.ErrActionRetry
	}
	return parallel.ExecuteLabeled(tasks, 2, retryPolicy)
}

func getRestartTasks(run model.Run, _ *CtrlClients) ([]parallel.LabeledTask, error) {
	logtrace.LogWithFunctionName()
	var controllers []*model.Component
	var err error

	scenarioCounter = (scenarioCounter + 1) % 7
	scenario := scenarioCounter + 1

	var result []parallel.LabeledTask

	if scenario&0b001 > 0 {
		controllers, err = chaos.SelectRandom(run, ".ctrl", chaos.RandomOfTotal())
		if err != nil {
			return nil, err
		}
		for _, controller := range controllers {
			result = append(result, parallel.TaskWithLabel("restart.ctrl", fmt.Sprintf("restart controller %s", controller.Id), func() error {
				return chaos.RestartSelected(run, 1, controller)
			}))
		}
	}

	var routers []*model.Component
	if scenario&0b010 > 0 {
		routers, err = chaos.SelectRandom(run, ".router", chaos.PercentageRange(10, 75))
		if err != nil {
			return nil, err
		}
		for _, router := range routers {
			result = append(result, parallel.TaskWithLabel("restart.router", fmt.Sprintf("restart router %s", router.Id), func() error {
				return chaos.RestartSelected(run, 1, router)
			}))
		}
	}

	return result, nil
}

func getRoles(n int) []string {
	logtrace.LogWithFunctionName()
	roles := getRoleAttributes(n)
	for i, role := range roles {
		roles[i] = "#" + role
	}
	return roles
}

func getRoleAttributes(n int) []string {
	logtrace.LogWithFunctionName()
	attr := map[string]struct{}{}
	count := rand.Intn(n) + 1
	for i := 0; i < count; i++ {
		attr[fmt.Sprintf("role-%v", rand.Intn(10))] = struct{}{}
	}

	var result []string
	for k := range attr {
		result = append(result, k)
	}
	return result
}

func getRoleAttributesAsAttrPtr(n int) *rest_model.Attributes {
	logtrace.LogWithFunctionName()
	result := getRoleAttributes(n)
	return (*rest_model.Attributes)(&result)
}

func newId() *string {
	logtrace.LogWithFunctionName()
	id := uuid.NewString()
	return &id
}

func newBoolPtr() *bool {
	logtrace.LogWithFunctionName()
	b := rand.Int()%2 == 0
	return &b
}

type taskGenerationContext struct {
	ctrls *CtrlClients

	configTypes []*rest_model.ConfigTypeDetail
	configs     []*rest_model.ConfigDetail
	services    []*rest_model.ServiceDetail

	configTypesDeleted map[string]struct{}
	configsDeleted     map[string]struct{}
	servicesDeleted    map[string]struct{}

	tasks     []parallel.LabeledTask
	lastTasks []parallel.LabeledTask

	configIdx int

	err error
}

func (self *taskGenerationContext) loadEntities() {
	logtrace.LogWithFunctionName()
	self.configTypes, self.err = models.ListConfigTypes(self.ctrls.getRandomCtrl(), `not (name contains ".v1" or id = "host.v2") limit none`, 15*time.Second)
	if self.err != nil {
		return
	}
	chaos.Randomize(self.configTypes)

	self.configs, self.err = models.ListConfigs(self.ctrls.getRandomCtrl(), "limit none", 15*time.Second)
	if self.err != nil {
		return
	}
	chaos.Randomize(self.configs)

	self.services, self.err = models.ListServices(self.ctrls.getRandomCtrl(), "limit none", 15*time.Second)
	if self.err != nil {
		return
	}
	chaos.Randomize(self.services)
}

func (self *taskGenerationContext) getConfigTypeId() string {
	logtrace.LogWithFunctionName()
	if len(self.configTypes)-len(self.configTypesDeleted) < 1 {
		return ""
	}

	for {
		idx := rand.Intn(len(self.configTypes))
		configType := self.configTypes[idx]
		if _, deleted := self.configTypesDeleted[*configType.ID]; !deleted {
			return *configType.ID
		}
	}
}

func (self *taskGenerationContext) getTwoValidConfigs() []string {
	logtrace.LogWithFunctionName()
	if len(self.configs)-len(self.configsDeleted) < 1 {
		return nil
	}
	var first *rest_model.ConfigDetail
	for {
		if self.configIdx >= len(self.configs) {
			self.configIdx = 0
		}
		next := self.configs[self.configIdx]
		self.configIdx++
		if _, deleted := self.configsDeleted[*next.ID]; deleted {
			continue
		}
		if first == nil {
			first = next
			continue
		}
		if first == next {
			return []string{*first.ID}
		}
		if *first.ConfigTypeID == *next.ConfigTypeID {
			continue
		}
		return []string{*first.ID, *next.ID}
	}
}

func (self *taskGenerationContext) generateConfigTypeTasks() {
	logtrace.LogWithFunctionName()
	if self.err != nil {
		return
	}

	if scenarioCounter%5 == 0 { // only add/remove config types every 5th iteration
		for i := 0; i < min(2, len(self.configTypes)); i++ {
			entityId := *self.configTypes[i].ID
			self.configTypesDeleted[entityId] = struct{}{}
			self.lastTasks = append(self.lastTasks, parallel.TaskWithLabel("delete.config-type", fmt.Sprintf("delete config type %s", entityId), func() error {
				return models.DeleteConfigType(self.ctrls.getRandomCtrl(), entityId, 15*time.Second)
			}))
		}
	}

	for i := 2; i < min(7, len(self.configTypes)); i++ {
		entity := self.configTypes[i]
		entityId := *entity.ID
		self.tasks = append(self.tasks, parallel.TaskWithLabel("modify.config-type", fmt.Sprintf("modify config type %s", entityId), func() error {
			entity.Name = newId()
			return models.UpdateConfigTypeFromDetail(self.ctrls.getRandomCtrl(), entity, 15*time.Second)
		}))
	}

	createConfigTypesCount := 15 - (len(self.configTypes) - len(self.configTypesDeleted)) // target 25 configs available
	for i := 0; i < createConfigTypesCount; i++ {
		self.tasks = append(self.tasks, createNewConfigType(self.ctrls.getRandomCtrl()))
	}
}

func (self *taskGenerationContext) generateConfigTasks() {
	logtrace.LogWithFunctionName()
	if self.err != nil {
		return
	}

	if scenarioCounter%3 == 0 && len(self.configs) > 2 { // only delete configs every third iteration
		for i := 0; i < 2; i++ {
			entityId := *self.configs[i].ID
			self.lastTasks = append(self.lastTasks, parallel.TaskWithLabel("delete.config", fmt.Sprintf("delete config %s", entityId), func() error {
				return models.DeleteConfig(self.ctrls.getRandomCtrl(), entityId, 15*time.Second)
			}))
		}
	}

	// delete any configs used by config types to be deleted
	if len(self.configTypesDeleted) > 0 {
		for _, config := range self.configs {
			if _, deleted := self.configsDeleted[*config.ID]; deleted {
				continue
			}
			if _, deleted := self.configTypesDeleted[*config.ConfigTypeID]; deleted {
				entityId := *config.ID
				self.configsDeleted[entityId] = struct{}{}
				self.lastTasks = append(self.lastTasks, parallel.TaskWithLabel("delete.config", fmt.Sprintf("delete config %s", entityId), func() error {
					return models.DeleteConfig(self.ctrls.getRandomCtrl(), entityId, 15*time.Second)
				}))
			}
		}
	}

	for i := 2; i < min(7, len(self.configs)); i++ {
		entityId := *self.configs[i].ID
		self.tasks = append(self.tasks, parallel.TaskWithLabel("modify.config", fmt.Sprintf("modify config %s", entityId), func() error {
			entity := self.configs[i]
			entity.Name = newId()
			entity.Data = map[string]interface{}{
				"hostname": fmt.Sprintf("https://%s.com", uuid.NewString()),
				"protocol": func() string {
					if rand.Int()%2 == 0 {
						return "tcp"
					}
					return "udp"
				}(),
				"port": rand.Intn(32000),
			}
			return models.UpdateConfigFromDetail(self.ctrls.getRandomCtrl(), entity, 15*time.Second)
		}))
	}

	if len(self.configTypes) > 0 {
		createConfigCount := 25 - (len(self.configs) - len(self.configsDeleted)) // target 25 configs available
		for i := 0; i < createConfigCount; i++ {
			self.tasks = append(self.tasks, createNewConfig(self.ctrls.getRandomCtrl(), self.getConfigTypeId()))
		}
	}
}

func (self *taskGenerationContext) generatePostureCheckTasks() {
	logtrace.LogWithFunctionName()
	if self.err != nil {
		return
	}

	if scenarioCounter%3 == 0 && len(self.configs) > 2 { // only delete configs every third iteration
		for i := 0; i < 2; i++ {
			entityId := *self.configs[i].ID
			self.lastTasks = append(self.lastTasks, parallel.TaskWithLabel("delete.config", fmt.Sprintf("delete config %s", entityId), func() error {
				return models.DeleteConfig(self.ctrls.getRandomCtrl(), entityId, 15*time.Second)
			}))
		}
	}

	// delete any configs used by config types to be deleted
	if len(self.configTypesDeleted) > 0 {
		for _, config := range self.configs {
			if _, deleted := self.configsDeleted[*config.ID]; deleted {
				continue
			}
			if _, deleted := self.configTypesDeleted[*config.ConfigTypeID]; deleted {
				entityId := *config.ID
				self.configsDeleted[entityId] = struct{}{}
				self.lastTasks = append(self.lastTasks, parallel.TaskWithLabel("delete.config", fmt.Sprintf("delete config %s", entityId), func() error {
					return models.DeleteConfig(self.ctrls.getRandomCtrl(), entityId, 15*time.Second)
				}))
			}
		}
	}

	for i := 2; i < min(7, len(self.configs)); i++ {
		entityId := *self.configs[i].ID
		self.tasks = append(self.tasks, parallel.TaskWithLabel("modify.config", fmt.Sprintf("modify config %s", entityId), func() error {
			entity := self.configs[i]
			entity.Name = newId()
			entity.Data = map[string]interface{}{
				"hostname": fmt.Sprintf("https://%s.com", uuid.NewString()),
				"protocol": func() string {
					if rand.Int()%2 == 0 {
						return "tcp"
					}
					return "udp"
				}(),
				"port": rand.Intn(32000),
			}
			return models.UpdateConfigFromDetail(self.ctrls.getRandomCtrl(), entity, 15*time.Second)
		}))
	}

	if len(self.configTypes) > 0 {
		createConfigCount := 25 - (len(self.configs) - len(self.configsDeleted)) // target 25 configs available
		for i := 0; i < createConfigCount; i++ {
			self.tasks = append(self.tasks, createNewConfig(self.ctrls.getRandomCtrl(), self.getConfigTypeId()))
		}
	}
}

func (self *taskGenerationContext) generateServiceTasks() {
	logtrace.LogWithFunctionName()
	if self.err != nil {
		return
	}

	for i := 0; i < min(5, len(self.services)); i++ {
		entityId := *self.services[i].ID
		self.servicesDeleted[entityId] = struct{}{}
		self.tasks = append(self.tasks, parallel.TaskWithLabel("delete.service", fmt.Sprintf("delete service %s", entityId), func() error {
			return models.DeleteService(self.ctrls.getRandomCtrl(), entityId, 15*time.Second)
		}))
	}

	modifyServiceTargetIndex := min(10, len(self.services))
	if len(self.configsDeleted) > 0 {
		modifyServiceTargetIndex = len(self.services)
	}
	for i := 5; i < modifyServiceTargetIndex; i++ {
		service := self.services[i]
		doModify := true
		if len(self.configsDeleted) > 0 {
			doModify = false
			for _, configId := range service.Configs {
				if _, deleted := self.configsDeleted[configId]; deleted {
					doModify = true
					break
				}
			}
		}
		if doModify {
			self.tasks = append(self.tasks, parallel.TaskWithLabel("modify.service", fmt.Sprintf("modify service %s", *service.ID), func() error {
				service.RoleAttributes = getRoleAttributesAsAttrPtr(3)
				service.Name = newId()
				service.Configs = self.getTwoValidConfigs()
				return models.UpdateServiceFromDetail(self.ctrls.getRandomCtrl(), service, 15*time.Second)
			}))
		}
	}

	createServicesTarget := 100 - (len(self.services) - len(self.servicesDeleted))
	for i := 0; i < createServicesTarget; i++ {
		self.tasks = append(self.tasks, createNewService(self.ctrls.getRandomCtrl(), self.getTwoValidConfigs()))
	}
}

func (self *taskGenerationContext) getResults() ([]parallel.LabeledTask, []parallel.LabeledTask, error) {
	logtrace.LogWithFunctionName()
	if self.err != nil {
		return nil, nil, self.err
	}

	// need to delete configs first, then config types
	slices.Reverse(self.lastTasks)
	return self.tasks, self.lastTasks, nil
}

func getServiceAndConfigChaosTasks(_ model.Run, ctrls *CtrlClients) ([]parallel.LabeledTask, []parallel.LabeledTask, error) {
	logtrace.LogWithFunctionName()
	ctx := &taskGenerationContext{
		ctrls:              ctrls,
		configTypesDeleted: map[string]struct{}{},
		configsDeleted:     map[string]struct{}{},
		servicesDeleted:    map[string]struct{}{},
	}

	ctx.loadEntities()
	ctx.generateConfigTypeTasks()
	ctx.generateConfigTasks()
	ctx.generatePostureCheckTasks()
	ctx.generateServiceTasks()

	return ctx.getResults()
}

func getIdentityChaosTasks(r model.Run, ctrls *CtrlClients) ([]parallel.LabeledTask, error) {
	logtrace.LogWithFunctionName()
	tunnelerCount := len(r.GetModel().SelectComponents("tunneler"))
	entities, err := models.ListIdentities(ctrls.getRandomCtrl(), "not isAdmin limit none", 15*time.Second)
	if err != nil {
		return nil, err
	}
	chaos.Randomize(entities)

	var result []parallel.LabeledTask

	var i int
	for len(result) < 5+(len(entities)-tunnelerCount-100) {
		entityId := *entities[i].ID
		if entities[i].Type.Name != "Router" {
			result = append(result, parallel.TaskWithLabel("delete.identity", fmt.Sprintf("delete identity %s", entityId), func() error {
				return models.DeleteIdentity(ctrls.getRandomCtrl(), entityId, 15*time.Second)
			}))
		}
		i++
	}

	for len(result) < 10 {
		entity := entities[i]
		result = append(result, parallel.TaskWithLabel("modify.identity", fmt.Sprintf("modify identity %s", *entity.ID), func() error {
			entity.RoleAttributes = getRoleAttributesAsAttrPtr(3)
			if entity.Type.Name != "Router" {
				entity.Name = newId()
			}
			return models.UpdateIdentityFromDetail(ctrls.getRandomCtrl(), entity, 15*time.Second)
		}))
		i++
	}

	for i := 0; i < 5; i++ {
		result = append(result, createNewIdentity(ctrls.getRandomCtrl()))
	}

	return result, nil
}

func getPostureTasks(r model.Run, ctrls *CtrlClients) ([]parallel.LabeledTask, error) {
	logtrace.LogWithFunctionName()
	entities, err := models.ListPostureChecks(ctrls.getRandomCtrl(), "limit none", 15*time.Second)
	if err != nil {
		return nil, err
	}
	chaos.Randomize(entities)

	var result []parallel.LabeledTask

	var i int
	for len(result) < 5+(len(entities)-100) {
		entityId := *entities[i].ID()
		result = append(result, parallel.TaskWithLabel("delete.posture-check", fmt.Sprintf("delete posture check %s", entityId), func() error {
			return models.DeletePostureCheck(ctrls.getRandomCtrl(), entityId, 15*time.Second)
		}))
		i++
	}

	for len(result) < min(10, len(entities)) {
		entity := entities[i]
		entity.SetName(newId())
		entity.SetRoleAttributes(getRoleAttributesAsAttrPtr(3))

		switch p := entity.(type) {
		case *rest_model.PostureCheckDomainDetail:
			p.Domains = []string{uuid.NewString(), uuid.NewString()}
		case *rest_model.PostureCheckMacAddressDetail:
			p.MacAddresses = []string{uuid.NewString(), uuid.NewString()}
		case *rest_model.PostureCheckMfaDetail:
			p.IgnoreLegacyEndpoints = *newBoolPtr()
			p.PromptOnUnlock = *newBoolPtr()
			p.PromptOnWake = *newBoolPtr()
			p.TimeoutSeconds = int64(rand.Intn(1000))
		case *rest_model.PostureCheckOperatingSystemDetail:
			p.OperatingSystems = getRandomOperatingSystems()
		case *rest_model.PostureCheckProcessDetail:
			p.Process = getRandomProcess()
		case *rest_model.PostureCheckProcessMultiDetail:
			p.Semantic = ptrutil.Ptr(getRandomSemantic())
			p.Processes = getRandomProcessMultis()
		default:
			return nil, fmt.Errorf("unhandled posture check type: %T", p)
		}

		result = append(result, parallel.TaskWithLabel("modify.posture-check", fmt.Sprintf("modify %s posture-check %s", entity.TypeID(), *entity.ID()), func() error {
			return models.UpdatePostureCheckFromDetail(ctrls.getRandomCtrl(), entity, 15*time.Second)
		}))
		i++
	}

	for i := 0; i < 55-len(entities); i++ {
		result = append(result, createNewPostureCheck(ctrls.getRandomCtrl()))
	}

	return result, nil
}

func getRandomSemantic() rest_model.Semantic {
	logtrace.LogWithFunctionName()
	if rand.Int()%2 == 0 {
		return rest_model.SemanticAnyOf
	}
	return rest_model.SemanticAllOf
}

func getRandomOperatingSystems() []*rest_model.OperatingSystem {
	logtrace.LogWithFunctionName()
	return getRandom(1, 3, getRandomOperatingSystem)
}

func getRandomOperatingSystem() *rest_model.OperatingSystem {
	logtrace.LogWithFunctionName()
	return &rest_model.OperatingSystem{
		Type:     ptrutil.Ptr(getRandomOsType()),
		Versions: getRandom(1, 3, getRandomVersion),
	}
}

func getRandomVersion() string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("%d.%d.%d", rand.Intn(100), rand.Intn(100), rand.Intn(100))
}

func getRandomProcessMultis() []*rest_model.ProcessMulti {
	logtrace.LogWithFunctionName()
	return getRandom(1, 3, getRandomProcessMulti)
}

func getRandomProcessMulti() *rest_model.ProcessMulti {
	logtrace.LogWithFunctionName()
	return &rest_model.ProcessMulti{
		Hashes:             []string{uuid.NewString(), uuid.NewString()},
		OsType:             ptrutil.Ptr(getRandomOsType()),
		Path:               ptrutil.Ptr(uuid.NewString()),
		SignerFingerprints: []string{uuid.NewString(), uuid.NewString()},
	}
}

func getRandomProcess() *rest_model.Process {
	logtrace.LogWithFunctionName()
	return &rest_model.Process{
		Hashes:            []string{uuid.NewString(), uuid.NewString()},
		OsType:            ptrutil.Ptr(getRandomOsType()),
		Path:              ptrutil.Ptr(uuid.NewString()),
		SignerFingerprint: uuid.NewString(),
	}
}

func getRandom[T any](min, max int, f func() T) []T {
	logtrace.LogWithFunctionName()
	var result []T
	count := min
	if max > min {
		min += rand.Intn(max - min)
	}
	for i := 0; i < count; i++ {
		result = append(result, f())
	}
	return result
}

var osTypes = []rest_model.OsType{
	rest_model.OsTypeLinux,
	rest_model.OsTypeWindows,
	rest_model.OsTypeMacOS,
	rest_model.OsTypeIOS,
	rest_model.OsTypeAndroid,
	rest_model.OsTypeWindowsServer,
}

func getRandomOsType() rest_model.OsType {
	logtrace.LogWithFunctionName()
	return osTypes[rand.Intn(len(osTypes))]
}

func createNewPostureCheck(ctrl *zitirest.Clients) parallel.LabeledTask {
	logtrace.LogWithFunctionName()
	var create rest_model.PostureCheckCreate

	switch rand.Intn(6) {
	case 0:
		create = &rest_model.PostureCheckDomainCreate{
			Domains: []string{uuid.NewString(), uuid.NewString()},
		}

	case 1:
		create = &rest_model.PostureCheckMacAddressCreate{
			MacAddresses: getRandom(1, 3, uuid.NewString),
		}

	case 2:
		mfaCreate := &rest_model.PostureCheckMfaCreate{}
		mfaCreate.IgnoreLegacyEndpoints = *newBoolPtr()
		mfaCreate.PromptOnUnlock = *newBoolPtr()
		mfaCreate.PromptOnWake = *newBoolPtr()
		mfaCreate.TimeoutSeconds = int64(rand.Intn(1000))
		create = mfaCreate
	case 3:
		create = &rest_model.PostureCheckOperatingSystemCreate{
			OperatingSystems: getRandomOperatingSystems(),
		}
	case 4:
		create = &rest_model.PostureCheckProcessCreate{
			Process: getRandomProcess(),
		}
	case 5:
		create = &rest_model.PostureCheckProcessMultiCreate{
			Semantic:  ptrutil.Ptr(getRandomSemantic()),
			Processes: getRandomProcessMultis(),
		}
	default:
		panic("programming error")
	}

	create.SetName(newId())
	create.SetRoleAttributes(getRoleAttributesAsAttrPtr(3))

	return parallel.TaskWithLabel("create.posture-check", fmt.Sprintf("create %s posture check", create.TypeID()), func() error {
		return models.CreatePostureCheck(ctrl, create, 15*time.Second)
	})
}

func getServicePolicyChaosTasks(_ model.Run, ctrls *CtrlClients) ([]parallel.LabeledTask, error) {
	logtrace.LogWithFunctionName()
	entities, err := models.ListServicePolicies(ctrls.getRandomCtrl(), "limit none", 15*time.Second)
	if err != nil {
		return nil, err
	}
	chaos.Randomize(entities)

	var result []parallel.LabeledTask

	for i := 0; i < 5; i++ {
		result = append(result, parallel.TaskWithLabel("delete.service-policy", fmt.Sprintf("delete service policy %s", *entities[i].ID), func() error {
			return models.DeleteServicePolicy(ctrls.getRandomCtrl(), *entities[i].ID, 15*time.Second)
		}))
	}

	for i := 5; i < 10; i++ {
		result = append(result, parallel.TaskWithLabel("modify.service-policy", fmt.Sprintf("modify service policy %s", *entities[i].ID), func() error {
			entity := entities[i]
			entity.IdentityRoles = getRoles(3)
			entity.ServiceRoles = getRoles(3)
			entity.PostureCheckRoles = getRoles(3)
			entity.Name = newId()
			return models.UpdateServicePolicyFromDetail(ctrls.getRandomCtrl(), entity, 15*time.Second)
		}))
	}

	for i := 0; i < 5; i++ {
		result = append(result, createNewServicePolicy(ctrls.getRandomCtrl()))
	}

	return result, nil
}

func createNewService(ctrl *zitirest.Clients, configs []string) parallel.LabeledTask {
	logtrace.LogWithFunctionName()
	return parallel.TaskWithLabel("create.service", "create new service", func() error {
		svc := &rest_model.ServiceCreate{
			Configs:            configs,
			EncryptionRequired: newBoolPtr(),
			Name:               newId(),
			RoleAttributes:     getRoleAttributes(3),
			TerminatorStrategy: "smartrouting",
		}
		return models.CreateService(ctrl, svc, 15*time.Second)
	})
}

func createNewConfigType(ctrl *zitirest.Clients) parallel.LabeledTask {
	logtrace.LogWithFunctionName()
	return parallel.TaskWithLabel("create.config-type", "create new config type", func() error {
		entity := &rest_model.ConfigTypeCreate{
			Name: newId(),
			Schema: map[string]interface{}{
				"$id":                  "https://edge.openziti.org/schemas/test.config.json",
				"type":                 "object",
				"additionalProperties": false,
				"required": []interface{}{
					"hostname",
					"port",
				},
				"properties": map[string]interface{}{
					"protocol": map[string]interface{}{
						"type": []interface{}{
							"string",
							"null",
						},
						"enum": []interface{}{
							"tcp",
							"udp",
						},
					},
					"hostname": map[string]interface{}{
						"type": "string",
					},
					"port": map[string]interface{}{
						"type":    "integer",
						"minimum": float64(0),
						"maximum": float64(math.MaxUint16),
					},
				},
			},
		}
		return models.CreateConfigType(ctrl, entity, 15*time.Second)
	})
}

func createNewConfig(ctrl *zitirest.Clients, configTypeId string) parallel.LabeledTask {
	logtrace.LogWithFunctionName()
	return parallel.TaskWithLabel("create.config", "create new config", func() error {
		entity := &rest_model.ConfigCreate{
			Name:         newId(),
			ConfigTypeID: &configTypeId,
			Data: map[string]interface{}{
				"hostname": fmt.Sprintf("https://%s.com", uuid.NewString()),
				"protocol": func() string {
					if rand.Int()%2 == 0 {
						return "tcp"
					}
					return "udp"
				}(),
				"port": rand.Intn(32000),
			},
		}
		return models.CreateConfig(ctrl, entity, 15*time.Second)
	})
}

func createNewIdentity(ctrl *zitirest.Clients) parallel.LabeledTask {
	logtrace.LogWithFunctionName()
	isAdmin := false
	identityType := rest_model.IdentityTypeDefault
	return parallel.TaskWithLabel("create.identity", "create new identity", func() error {
		svc := &rest_model.IdentityCreate{
			DefaultHostingCost:        nil,
			DefaultHostingPrecedence:  "",
			IsAdmin:                   &isAdmin,
			Name:                      newId(),
			RoleAttributes:            getRoleAttributesAsAttrPtr(3),
			ServiceHostingCosts:       nil,
			ServiceHostingPrecedences: nil,
			Tags:                      nil,
			Type:                      &identityType,
		}
		return models.CreateIdentity(ctrl, svc, 15*time.Second)
	})
}

func createNewServicePolicy(ctrl *zitirest.Clients) parallel.LabeledTask {
	logtrace.LogWithFunctionName()
	return parallel.TaskWithLabel("create.service-policy", "create new service policy", func() error {
		policyType := rest_model.DialBindDial
		if rand.Int()%2 == 0 {
			policyType = rest_model.DialBindBind
		}
		entity := &rest_model.ServicePolicyCreate{
			Name:              newId(),
			IdentityRoles:     getRoles(3),
			PostureCheckRoles: getRoles(3),
			Semantic:          ptrutil.Ptr(getRandomSemantic()),
			ServiceRoles:      getRoles(3),
			Type:              &policyType,
		}
		return models.CreateServicePolicy(ctrl, entity, 15*time.Second)
	})
}

func validateRouterDataModel(run model.Run) error {
	logtrace.LogWithFunctionName()
	ctrls := run.GetModel().SelectComponents(".ctrl")
	errC := make(chan error, len(ctrls))
	deadline := time.Now().Add(15 * time.Minute)
	for _, ctrl := range ctrls {
		ctrlComponent := ctrl
		go validateRouterDataModelForCtrlWithChan(run, ctrlComponent, deadline, errC)
	}

	for i := 0; i < len(ctrls); i++ {
		err := <-errC
		if err != nil {
			return err
		}
	}

	return nil
}

func validateRouterDataModelForCtrlWithChan(run model.Run, c *model.Component, deadline time.Time, errC chan<- error) {
	logtrace.LogWithFunctionName()
	errC <- validateRouterDataModelForCtrl(run, c, deadline)
}

func validateRouterDataModelForCtrl(run model.Run, c *model.Component, deadline time.Time) error {
	logtrace.LogWithFunctionName()
	clients, err := chaos.EnsureLoggedIntoCtrl(run, c, time.Minute)
	if err != nil {
		return err
	}

	start := time.Now()

	logger := pfxlog.Logger().WithField("ctrl", c.Id)

	for {
		count, err := validateRouterDataModelForCtrlOnce(c.Id, clients)
		if err == nil {
			return nil
		}

		if time.Now().After(deadline) {
			return err
		}

		logger.Infof("current count of router data model errors: %v, elapsed time: %v", count, time.Since(start))
		time.Sleep(15 * time.Second)

		clients, err = chaos.EnsureLoggedIntoCtrl(run, c, time.Minute)
		if err != nil {
			return err
		}
	}
}

func validateRouterDataModelForCtrlOnce(id string, clients *zitirest.Clients) (int, error) {
	logtrace.LogWithFunctionName()
	logger := pfxlog.Logger().WithField("ctrl", id)

	closeNotify := make(chan struct{})
	eventNotify := make(chan *mgmt_pb.RouterDataModelDetails, 1)

	handleSdkTerminatorResults := func(msg *channel.Message, _ channel.Channel) {
		detail := &mgmt_pb.RouterDataModelDetails{}
		if err := proto.Unmarshal(msg.Body, detail); err != nil {
			pfxlog.Logger().WithError(err).Error("unable to unmarshal router data model details")
			return
		}
		eventNotify <- detail
	}

	bindHandler := func(binding channel.Binding) error {
		binding.AddReceiveHandlerF(int32(mgmt_pb.ContentType_ValidateRouterDataModelResultType), handleSdkTerminatorResults)
		binding.AddCloseHandler(channel.CloseHandlerF(func(ch channel.Channel) {
			close(closeNotify)
		}))
		return nil
	}

	ch, err := clients.NewWsMgmtChannel(channel.BindHandlerF(bindHandler))
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = ch.Close()
	}()

	request := &mgmt_pb.ValidateRouterDataModelRequest{
		RouterFilter: "limit none",
		ValidateCtrl: true,
	}
	responseMsg, err := protobufs.MarshalTyped(request).WithTimeout(10 * time.Second).SendForReply(ch)

	response := &mgmt_pb.ValidateRouterDataModelResponse{}
	if err = protobufs.TypedResponse(response).Unmarshall(responseMsg, err); err != nil {
		return 0, err
	}

	if !response.Success {
		return 0, fmt.Errorf("failed to start router data model validation: %s", response.Message)
	}

	logger.Infof("started validation of %v components", response.ComponentCount)

	expected := response.ComponentCount

	invalid := 0
	for expected > 0 {
		select {
		case <-closeNotify:
			fmt.Printf("channel closed, exiting")
			return 0, errors.New("unexpected close of mgmt channel")
		case detail := <-eventNotify:
			if !detail.ValidateSuccess {
				invalid++
			}
			for _, errorDetails := range detail.Errors {
				fmt.Printf("\tdetail: %s\n", errorDetails)
			}
			expected--
		}
	}
	if invalid == 0 {
		logger.Infof("router data model validation of %v components successful", response.ComponentCount)
		return invalid, nil
	}
	return invalid, errors.New("errors found")
}
