package event

import (
	"fmt"
	"ztna-core/ztna/logtrace"
)

const (
	UsageEventsNs      = "fabric.usage"
	UsageEventsVersion = 2
)

type UsageEvent struct {
	Namespace        string            `json:"namespace"`
	Version          uint32            `json:"version"`
	EventType        string            `json:"event_type"`
	EventSrcId       string            `json:"event_src_id"`
	SourceId         string            `json:"source_id"`
	CircuitId        string            `json:"circuit_id"`
	Usage            uint64            `json:"usage"`
	IntervalStartUTC int64             `json:"interval_start_utc"`
	IntervalLength   uint64            `json:"interval_length"`
	Tags             map[string]string `json:"tags"`
}

func (event *UsageEvent) String() string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("%v source=%v session=%v usage=%v intervalStart=%v intervalLength=%v",
		event.EventType, event.SourceId, event.CircuitId, event.Usage, event.IntervalStartUTC, event.IntervalLength)
}

type UsageEventHandler interface {
	AcceptUsageEvent(event *UsageEvent)
}

type UsageEventV3 struct {
	Namespace        string            `json:"namespace"`
	Version          uint32            `json:"version"`
	EventSrcId       string            `json:"event_src_id"`
	SourceId         string            `json:"source_id"`
	CircuitId        string            `json:"circuit_id"`
	Usage            map[string]uint64 `json:"usage"`
	IntervalStartUTC int64             `json:"interval_start_utc"`
	IntervalLength   uint64            `json:"interval_length"`
	Tags             map[string]string `json:"tags"`
}

func (event *UsageEventV3) String() string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("source=%v session=%v usage=%v intervalStart=%v intervalLength=%v",
		event.SourceId, event.CircuitId, event.Usage, event.IntervalStartUTC, event.IntervalLength)
}

type UsageEventV3Handler interface {
	AcceptUsageEventV3(event *UsageEventV3)
}

type UsageEventV3HandlerWrapper interface {
	UsageEventV3Handler
	IsWrapping(value UsageEventV3Handler) bool
}
