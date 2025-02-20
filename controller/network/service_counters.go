package network

import (
	"fmt"
	"time"
	"ztna-core/ztna/logtrace"
)

type ServiceCounters interface {
	ServiceDialSuccess(serviceId, terminatorId string)
	ServiceDialFail(serviceId, terminatorId string)
	ServiceDialTimeout(serviceId, terminatorId string)
	ServiceDialOtherError(serviceId string)

	ServiceTerminatorTimeout(serviceId, terminatorId string)
	ServiceTerminatorConnectionRefused(serviceId, terminatorId string)
	ServiceInvalidTerminator(serviceId, terminatorId string)
	ServiceMisconfiguredTerminator(serviceId, terminatorId string)
}

func (network *Network) ServiceDialSuccess(serviceId, terminatorId string) {
	logtrace.LogWithFunctionName()
	combinedId := network.joinIds(serviceId, terminatorId)
	network.serviceDialSuccessCounter.Update(combinedId, time.Now(), 1)
}

func (network *Network) ServiceDialFail(serviceId, terminatorId string) {
	logtrace.LogWithFunctionName()
	combinedId := network.joinIds(serviceId, terminatorId)
	network.serviceDialFailCounter.Update(combinedId, time.Now(), 1)
}

func (network *Network) ServiceDialTimeout(serviceId, terminatorId string) {
	logtrace.LogWithFunctionName()
	combinedId := network.joinIds(serviceId, terminatorId)
	network.serviceDialTimeoutCounter.Update(combinedId, time.Now(), 1)
}

func (network *Network) ServiceDialOtherError(serviceId string) {
	logtrace.LogWithFunctionName()
	network.serviceDialOtherErrorCounter.Update(serviceId, time.Now(), 1)
}

func (network *Network) ServiceTerminatorTimeout(serviceId, terminatorId string) {
	logtrace.LogWithFunctionName()
	combinedId := network.joinIds(serviceId, terminatorId)
	network.serviceTerminatorTimeoutCounter.Update(combinedId, time.Now(), 1)
}
func (network *Network) ServiceTerminatorConnectionRefused(serviceId, terminatorId string) {
	logtrace.LogWithFunctionName()
	combinedId := network.joinIds(serviceId, terminatorId)
	network.serviceTerminatorConnectionRefusedCounter.Update(combinedId, time.Now(), 1)
}
func (network *Network) ServiceInvalidTerminator(serviceId, terminatorId string) {
	logtrace.LogWithFunctionName()
	combinedId := network.joinIds(serviceId, terminatorId)
	network.serviceInvalidTerminatorCounter.Update(combinedId, time.Now(), 1)
}

func (network *Network) ServiceMisconfiguredTerminator(serviceId, terminatorId string) {
	logtrace.LogWithFunctionName()
	combinedId := network.joinIds(serviceId, terminatorId)
	network.serviceMisconfiguredTerminatorCounter.Update(combinedId, time.Now(), 1)
}

func (network *Network) joinIds(serviceId, terminatorId string) string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("%v:%v", serviceId, terminatorId)
}
