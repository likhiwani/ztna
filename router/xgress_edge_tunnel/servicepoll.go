package xgress_edge_tunnel

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

import (
	"reflect"
	"sync"
	"time"

	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/tunnel/intercept"

	"ztna-core/sdk-golang/ziti"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/sirupsen/logrus"
)

func newServicePoller(fabricProvider *fabricProvider) *servicePoller {
	logtrace.LogWithFunctionName()
	result := &servicePoller{
		services:                cmap.New[*rest_model.ServiceDetail](),
		servicesLastUpdateToken: cmap.New[[]byte](),
		fabricProvider:          fabricProvider,
	}

	return result
}

type servicePoller struct {
	services                cmap.ConcurrentMap[string, *rest_model.ServiceDetail]
	serviceListener         *intercept.ServiceListener
	servicesLastUpdateToken cmap.ConcurrentMap[string, []byte]
	serviceListenerLock     sync.Mutex

	fabricProvider *fabricProvider
}

func (self *servicePoller) handleServiceListUpdate(ch channel.Channel, lastUpdateToken []byte, services []*rest_model.ServiceDetail) {
	logtrace.LogWithFunctionName()
	self.serviceListenerLock.Lock()
	defer self.serviceListenerLock.Unlock()

	if self.serviceListener == nil {
		logrus.Error("GOT SERVICE LIST BEFORE INITIALIZATION COMPLETE. SHOULD NOT HAPPEN!")
		return
	}

	logrus.Debugf("processing service updates with %v services", len(services))

	self.servicesLastUpdateToken.Set(ch.Id(), lastUpdateToken)

	idMap := make(map[string]*rest_model.ServiceDetail)
	for _, s := range services {
		idMap[*s.ID] = s
	}

	var toRemove []string

	// process Deletes
	self.services.IterCb(func(k string, svc *rest_model.ServiceDetail) {
		if _, found := idMap[*svc.ID]; !found {
			toRemove = append(toRemove, k)
			self.serviceListener.HandleServicesChange(ziti.ServiceRemoved, svc)
		}
	})

	for _, key := range toRemove {
		self.services.Remove(key)
	}

	// Adds and Updates
	for _, s := range services {
		self.services.Upsert(*s.ID, s, func(exist bool, valueInMap *rest_model.ServiceDetail, newValue *rest_model.ServiceDetail) *rest_model.ServiceDetail {
			if !exist {
				self.serviceListener.HandleServicesChange(ziti.ServiceAdded, s)
				return s
			}
			if !reflect.DeepEqual(valueInMap, s) {
				self.serviceListener.HandleServicesChange(ziti.ServiceChanged, s)
				return s
			} else {
				logrus.WithField("service", s.Name).Debug("no change detected in service definition")
			}
			return valueInMap
		})
	}
}

// TODO: just push updates down the control channel when necessary
func (self *servicePoller) pollServices(pollInterval time.Duration, notifyClose <-chan struct{}) {
	logtrace.LogWithFunctionName()
	if err := self.fabricProvider.authenticate(); err != nil {
		logrus.WithError(err).Fatal("xgress_edge_tunnel unable to authenticate to controller. " +
			"ensure tunneler mode is enabled for this router or disable tunnel listener. exiting ")
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	self.requestServiceListUpdate()

	for {
		select {
		case <-ticker.C:
			self.requestServiceListUpdate()
		case <-notifyClose:
			return
		}
	}
}

func (self *servicePoller) requestServiceListUpdate() {
	logtrace.LogWithFunctionName()
	ctrlCh := self.fabricProvider.factory.ctrls.AnyCtrlChannel()
	if ctrlCh != nil { // not currently connected to any controllers
		lastUpdateToken, _ := self.servicesLastUpdateToken.Get(ctrlCh.Id())
		self.fabricProvider.requestServiceList(ctrlCh, lastUpdateToken)
	} else {
		pfxlog.Logger().Warn("unable to request service list update, no controllers connected")
	}
}
