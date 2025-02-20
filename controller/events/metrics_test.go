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

package events

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"
	"ztna-core/ztna/controller/event"
	"ztna-core/ztna/logtrace"

	metrics2 "github.com/openziti/metrics"
	"github.com/stretchr/testify/require"
)

func Test_ExtractId(t *testing.T) {
	logtrace.LogWithFunctionName()
	name := "ctrl.3tOOkKfDn.tx.bytesrate"

	req := require.New(t)
	name, entityId := ExtractId(name, "ctrl.", 2)
	req.Equal(name, "ctrl.tx.bytesrate")
	req.Equal(entityId, "3tOOkKfDn")

	name = "ctrl.3tO.kKfDn.tx.bytesrate"
	name, entityId = ExtractId(name, "ctrl.", 2)
	req.Equal(name, "ctrl.tx.bytesrate")
	req.Equal(entityId, "3tO.kKfDn")

	name = "ctrl.3tO.kK.Dn.tx.bytesrate"
	name, entityId = ExtractId(name, "ctrl.", 2)
	req.Equal(name, "ctrl.tx.bytesrate")
	req.Equal(entityId, "3tO.kK.Dn")

	name = "ctrl..tO.kK.Dn.tx.bytesrate"
	name, entityId = ExtractId(name, "ctrl.", 2)
	req.Equal(name, "ctrl.tx.bytesrate")
	req.Equal(entityId, ".tO.kK.Dn")

	name = "ctrl..tO.kK.D..tx.bytesrate"
	name, entityId = ExtractId(name, "ctrl.", 2)
	req.Equal(name, "ctrl.tx.bytesrate")
	req.Equal(entityId, ".tO.kK.D.")
}

func Test_FilterMetrics(t *testing.T) {
	logtrace.LogWithFunctionName()
	req := require.New(t)

	closeNotify := make(chan struct{})
	defer close(closeNotify)
	dispatcher := NewDispatcher(closeNotify)
	dispatcher.ctrlId = "ctrl1"

	unfilteredEventC := make(chan *event.MetricsEvent, 1)
	adapter := dispatcher.NewFilteredMetricsAdapter(nil, nil, event.MetricsEventHandlerF(func(evt *event.MetricsEvent) {
		unfilteredEventC <- evt
	}))
	dispatcher.AddMetricsMessageHandler(adapter)

	filteredEventC := make(chan *event.MetricsEvent, 1)
	filter, err := regexp.Compile("foo.bar.(m1_rate|count)")
	req.NoError(err)
	adapter = dispatcher.NewFilteredMetricsAdapter(nil, filter, event.MetricsEventHandlerF(func(evt *event.MetricsEvent) {
		filteredEventC <- evt
	}))

	dispatcher.AddMetricsMessageHandler(adapter)

	go func() {
		registry := metrics2.NewRegistry("test", nil)
		meter := registry.Meter("foo.bar")
		meter.Mark(1)
		dispatcher.AcceptMetricsMsg(registry.Poll())
	}()

	var evt *event.MetricsEvent
	select {
	case evt = <-unfilteredEventC:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for unfiltered event")
	}

	req.Equal("foo.bar", evt.Metric)
	fmt.Printf("%+v\n", evt.Metrics)
	req.Equal(5, len(evt.Metrics))
	req.Equal(int64(1), evt.Metrics["count"])
	req.NotNil(evt.Metrics["mean_rate"])
	req.NotNil(evt.Metrics["m1_rate"])
	req.NotNil(evt.Metrics["m5_rate"])
	req.NotNil(evt.Metrics["m15_rate"])

	select {
	case evt = <-filteredEventC:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for filtered event")
	}

	req.Equal("foo.bar", evt.Metric)
	fmt.Printf("%+v\n", evt.Metrics)
	req.Equal(2, len(evt.Metrics))
	req.Equal(int64(1), evt.Metrics["count"])
	req.NotNil(evt.Metrics["m1_rate"])
}

func Test_MetricsFormat(t *testing.T) {
	logtrace.LogWithFunctionName()
	req := require.New(t)

	closeNotify := make(chan struct{})
	defer close(closeNotify)
	dispatcher := NewDispatcher(closeNotify)

	unfilteredEventC := make(chan *event.MetricsEvent, 1)
	adapter := dispatcher.NewFilteredMetricsAdapter(nil, nil, event.MetricsEventHandlerF(func(evt *event.MetricsEvent) {
		unfilteredEventC <- evt
	}))
	dispatcher.AddMetricsMessageHandler(adapter)

	go func() {
		registry := metrics2.NewRegistry("test", nil)
		meter := registry.Meter("foo.bar")
		time.Sleep(10 * time.Millisecond)
		meter.Mark(1)
		dispatcher.AcceptMetricsMsg(registry.Poll())
	}()

	var evt *event.MetricsEvent
	select {
	case evt = <-unfilteredEventC:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for unfiltered event")
	}

	req.Equal("foo.bar", evt.Metric)
	fmt.Printf("%+v\n", evt.Metrics)
	req.Equal(5, len(evt.Metrics))
	req.Equal(int64(1), evt.Metrics["count"])
	req.NotNil(evt.Metrics["mean_rate"])
	req.NotNil(evt.Metrics["m1_rate"])
	req.NotNil(evt.Metrics["m5_rate"])
	req.NotNil(evt.Metrics["m15_rate"])

	jsonEvent := (*JsonMetricsEvent)(evt)
	buf, err := jsonEvent.Format()
	req.NoError(err)

	jsonData := map[string]any{}
	req.NoError(json.Unmarshal(buf, &jsonData))

	req.Equal("metrics", jsonData["namespace"])
	req.Equal(float64(event.MetricsEventsVersion), jsonData["version"])
	req.Equal("meter", jsonData["metric_type"])
	req.Equal("foo.bar", jsonData["metric"])
	req.Equal("test", jsonData["source_id"])
	req.Equal(evt.SourceEventId, jsonData["source_event_id"])
	req.Equal(evt.Timestamp.Format(time.RFC3339Nano), jsonData["timestamp"])

	nested, ok := jsonData["metrics"]
	req.True(ok)
	nestedJson, ok := nested.(map[string]any)
	req.True(ok)

	req.Equal(float64(1), nestedJson["count"])
}
