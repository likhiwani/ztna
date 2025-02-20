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

package forwarder

import (
	"time"
	"ztna-core/ztna/common/inspect"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/common/trace"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/env"
	"ztna-core/ztna/router/xgress"
	"ztna-core/ztna/router/xlink"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/foundation/v2/info"
	"github.com/openziti/metrics"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Forwarder struct {
	circuits        *circuitTable
	destinations    *destinationTable
	faulter         FaultReceiver
	metricsRegistry metrics.UsageRegistry
	traceController trace.Controller
	Options         *Options
	CloseNotify     <-chan struct{}
}

type Destination interface {
	SendPayload(payload *xgress.Payload, timeout time.Duration, payloadType xgress.PayloadType) error
	SendAcknowledgement(acknowledgement *xgress.Acknowledgement) error
	SendControl(control *xgress.Control) error
	InspectCircuit(detail *inspect.CircuitInspectDetail)
}

type XgressDestination interface {
	Destination
	Unrouted()
	Start()
	IsTerminator() bool
	Label() string
	GetTimeOfLastRxFromLink() int64
}

func NewForwarder(metricsRegistry metrics.UsageRegistry, faulter FaultReceiver, options *Options, closeNotify <-chan struct{}) *Forwarder {
	logtrace.LogWithFunctionName()
	f := &Forwarder{
		circuits:        newCircuitTable(),
		destinations:    newDestinationTable(),
		faulter:         faulter,
		metricsRegistry: metricsRegistry,
		traceController: trace.NewController(closeNotify),
		Options:         options,
		CloseNotify:     closeNotify,
	}
	return f
}

func (forwarder *Forwarder) StartScanner(ctrls env.NetworkControllers) {
	logtrace.LogWithFunctionName()
	scanner := newScanner(ctrls, forwarder.Options, forwarder.CloseNotify)
	scanner.setCircuitTable(forwarder.circuits)

	if scanner.interval > 0 {
		go scanner.run()
	} else {
		logrus.Warnf("scanner disabled")
	}
}

func (forwarder *Forwarder) MetricsRegistry() metrics.UsageRegistry {
	logtrace.LogWithFunctionName()
	return forwarder.metricsRegistry
}

func (forwarder *Forwarder) TraceController() trace.Controller {
	logtrace.LogWithFunctionName()
	return forwarder.traceController
}

func (forwarder *Forwarder) RegisterDestination(circuitId string, address xgress.Address, destination Destination) {
	logtrace.LogWithFunctionName()
	forwarder.destinations.addDestination(address, destination)
	forwarder.destinations.linkDestinationToCircuit(circuitId, address)
}

func (forwarder *Forwarder) UnregisterDestinations(circuitId string) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger().WithField("circuitId", circuitId)
	if addresses, found := forwarder.destinations.getAddressesForCircuit(circuitId); found {
		for _, address := range addresses {
			if destination, found := forwarder.destinations.getDestination(address); found {
				log.Debugf("unregistering destination [@/%v] for circuit", address)
				forwarder.destinations.removeDestination(address)
				go destination.(XgressDestination).Unrouted()
			} else {
				log.Debugf("no destinations found for [@/%v] for circuit", address)
			}
		}
		forwarder.destinations.unlinkCircuit(circuitId)
	} else {
		log.Debug("found no addresses to unregister for circuit")
	}
}

func (forwarder *Forwarder) HasDestination(address xgress.Address) bool {
	logtrace.LogWithFunctionName()
	_, found := forwarder.destinations.getDestination(address)
	return found
}

func (forwarder *Forwarder) RegisterLink(link xlink.LinkDestination) error {
	logtrace.LogWithFunctionName()
	forwarder.destinations.addDestination(xgress.Address(link.Id()), link)
	return nil
}

func (forwarder *Forwarder) UnregisterLink(link xlink.LinkDestination) {
	logtrace.LogWithFunctionName()
	forwarder.destinations.removeDestinationIfMatches(xgress.Address(link.Id()), link)
}

func (forwarder *Forwarder) Route(ctrlId string, route *ctrl_pb.Route) error {
	logtrace.LogWithFunctionName()
	circuitId := route.CircuitId
	var circuitFt *forwardTable
	if ft, found := forwarder.circuits.getForwardTable(circuitId, true); found {
		circuitFt = ft
	} else {
		circuitFt = newForwardTable(ctrlId)
	}
	for _, forward := range route.Forwards {
		if !forwarder.HasDestination(xgress.Address(forward.DstAddress)) {
			if forward.DstType == ctrl_pb.DestType_Link {
				forwarder.faulter.NotifyInvalidLink(forward.DstAddress)
				return errors.Errorf("invalid link destination %v", forward.DstAddress)
			}
			if forward.DstType == ctrl_pb.DestType_End {
				return errors.Errorf("invalid egress destination %v", forward.DstAddress)
			}
			// It's an ingress destination, which isn't established until after routing has completed
		}
		circuitFt.setForwardAddress(xgress.Address(forward.SrcAddress), xgress.Address(forward.DstAddress))
	}
	forwarder.circuits.setForwardTable(circuitId, circuitFt)
	return nil
}

func (forwarder *Forwarder) Unroute(circuitId string, now bool) {
	logtrace.LogWithFunctionName()
	if now {
		forwarder.circuits.removeForwardTable(circuitId)
		forwarder.EndCircuit(circuitId)
	} else {
		go forwarder.unrouteTimeout(circuitId, forwarder.Options.XgressCloseCheckInterval)
	}
}

func (forwarder *Forwarder) EndCircuit(circuitId string) {
	logtrace.LogWithFunctionName()
	forwarder.UnregisterDestinations(circuitId)
}

func (forwarder *Forwarder) ForwardPayload(srcAddr xgress.Address, payload *xgress.Payload, timeout time.Duration) error {
	logtrace.LogWithFunctionName()
	return forwarder.forwardPayload(srcAddr, payload, true, timeout)
}

func (forwarder *Forwarder) RetransmitPayload(srcAddr xgress.Address, payload *xgress.Payload) error {
	logtrace.LogWithFunctionName()
	return forwarder.forwardPayload(srcAddr, payload, false, 0)
}

func (forwarder *Forwarder) forwardPayload(srcAddr xgress.Address, payload *xgress.Payload, markActive bool, timeout time.Duration) error {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(string(srcAddr))

	circuitId := payload.GetCircuitId()
	if forwardTable, found := forwarder.circuits.getForwardTable(circuitId, markActive); found {
		if dstAddr, found := forwardTable.getForwardAddress(srcAddr); found {
			if dst, found := forwarder.destinations.getDestination(dstAddr); found {
				payloadType := xgress.PayloadTypeXg
				if !markActive {
					payloadType = xgress.PayloadTypeRtx
				} else if timeout == 0 {
					payloadType = xgress.PayloadTypeFwd
				}
				if err := dst.SendPayload(payload, timeout, payloadType); err != nil {
					return err
				}
				log.WithFields(payload.GetLoggerFields()).Debugf("=> %s", string(dstAddr))
				return nil
			} else {
				return errors.Errorf("cannot forward payload, no destination for circuit=%v src=%v dst=%v", circuitId, srcAddr, dstAddr)
			}
		} else {
			return errors.Errorf("cannot forward payload, no destination address for circuit=%v src=%v", circuitId, srcAddr)
		}
	} else {
		return errors.Errorf("cannot forward payload, no forward table for circuit=%v src=%v", circuitId, srcAddr)
	}
}

func (forwarder *Forwarder) ForwardAcknowledgement(srcAddr xgress.Address, acknowledgement *xgress.Acknowledgement) error {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(string(srcAddr))

	circuitId := acknowledgement.CircuitId
	if forwardTable, found := forwarder.circuits.getForwardTable(circuitId, true); found {
		if dstAddr, found := forwardTable.getForwardAddress(srcAddr); found {
			if dst, found := forwarder.destinations.getDestination(dstAddr); found {
				if err := dst.SendAcknowledgement(acknowledgement); err != nil {
					return err
				}
				log.Debugf("=> %s", string(dstAddr))
				return nil

			} else {
				return errors.Errorf("cannot acknowledge, no destination for circuit=%v src=%v dst=%v", circuitId, srcAddr, dstAddr)
			}

		} else {
			return errors.Errorf("cannot acknowledge, no destination address for circuit=%v src=%v", circuitId, srcAddr)
		}

	} else {
		return errors.Errorf("cannot acknowledge, no forward table for circuit=%v src=%v", circuitId, srcAddr)
	}
}

func (forwarder *Forwarder) ForwardControl(srcAddr xgress.Address, control *xgress.Control) error {
	logtrace.LogWithFunctionName()
	circuitId := control.CircuitId
	log := pfxlog.ContextLogger(string(srcAddr)).WithField("circuitId", circuitId)

	var err error

	if forwardTable, found := forwarder.circuits.getForwardTable(circuitId, true); found {
		if dstAddr, found := forwardTable.getForwardAddress(srcAddr); found {
			if dst, found := forwarder.destinations.getDestination(dstAddr); found {
				if control.IsTypeTraceRoute() {
					hops := control.DecrementAndGetHop()
					if hops == 0 {
						resp := control.CreateTraceResponse("forwarder", forwarder.metricsRegistry.SourceId())
						return forwarder.ForwardControl(dstAddr, resp)
					}
				}
				err = dst.SendControl(control)
				log.Debugf("=> %s", string(dstAddr))
			} else {
				err = errors.Errorf("cannot forward control, no destination for circuit=%v src=%v dst=%v", circuitId, srcAddr, dstAddr)
			}
		} else {
			err = errors.Errorf("cannot forward control, no destination address for circuit=%v src=%v", circuitId, srcAddr)
		}
	} else {
		err = errors.Errorf("cannot forward control, no forward table for circuit=%v src=%v", circuitId, srcAddr)
	}

	if err != nil && control.IsTypeTraceRoute() {
		resp := control.CreateTraceResponse("forwarder", forwarder.metricsRegistry.SourceId())
		resp.Headers.PutStringHeader(xgress.ControlError, err.Error())
		if dst, found := forwarder.destinations.getDestination(srcAddr); found {
			if fwdErr := dst.SendControl(resp); fwdErr != nil {
				log.WithError(errorz.MultipleErrors{err, fwdErr}).Error("error sending trace error response")
			}
		} else {
			log.WithError(err).Error("unable to send trace error response as destination for source not found")
		}
	}

	return err
}

func (forwarder *Forwarder) ReportForwardingFault(circuitId string, ctrlId string) {
	logtrace.LogWithFunctionName()
	if forwarder.faulter != nil {
		forwarder.faulter.Report(circuitId, ctrlId)
	} else {
		logrus.Error("nil faulter, cannot accept forwarding fault report")
	}
}

func (forwarder *Forwarder) Debug() string {
	logtrace.LogWithFunctionName()
	return forwarder.circuits.debug() + forwarder.destinations.debug()
}

// unrouteTimeout implements a goroutine to manage route timeout processing. Once a timeout processor has been launched
// for a circuit, it will be checked repeatedly, looking to see if the circuit has crossed the inactivity threshold.
// Once it crosses the inactivity threshold, it gets removed.
func (forwarder *Forwarder) unrouteTimeout(circuitId string, interval time.Duration) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger("c/" + circuitId)
	log.Debug("scheduled")
	defer log.Debug("timeout")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if dest := forwarder.getXgressForCircuit(circuitId); dest != nil {
				elapsedDelta := info.NowInMilliseconds() - dest.GetTimeOfLastRxFromLink()
				if (time.Duration(elapsedDelta) * time.Millisecond) >= interval {
					forwarder.circuits.removeForwardTable(circuitId)
					forwarder.EndCircuit(circuitId)
					return
				}
			} else {
				forwarder.circuits.removeForwardTable(circuitId)
				forwarder.EndCircuit(circuitId)
				return
			}
		case <-forwarder.CloseNotify:
			return
		}
	}
}

func (forwarder *Forwarder) getXgressForCircuit(circuitId string) XgressDestination {
	logtrace.LogWithFunctionName()
	if addresses, found := forwarder.destinations.getAddressesForCircuit(circuitId); found {
		for _, address := range addresses {
			if destination, found := forwarder.destinations.getDestination(address); found {
				return destination.(XgressDestination)
			}
		}
	}
	return nil
}

func (forwarder *Forwarder) InspectCircuit(circuitId string, getRelatedGoroutines bool) *inspect.CircuitInspectDetail {
	logtrace.LogWithFunctionName()
	if ft, found := forwarder.circuits.circuits.Get(circuitId); found {
		result := &inspect.CircuitInspectDetail{
			CircuitId:     circuitId,
			Forwards:      map[string]string{},
			XgressDetails: map[string]*inspect.XgressDetail{},
			LinkDetails:   map[string]*inspect.LinkInspectDetail{},
		}
		result.SetIncludeGoroutines(getRelatedGoroutines)

		ft.destinations.IterCb(func(key string, dest string) {
			result.Forwards[key] = dest
		})

		for _, addr := range result.Forwards {
			if dest, _ := forwarder.destinations.getDestination(xgress.Address(addr)); dest != nil {
				dest.InspectCircuit(result)
			}
		}
		return result
	}
	return nil
}
