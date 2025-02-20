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

package chaos

import (
	"fmt"
	"math/rand"
	"time"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/zitirest"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/model"
)

func StaticNumber(val int) func(int) int {
	logtrace.LogWithFunctionName()
	return func(int) int {
		return val
	}
}

func RandomInRange(min, max int) func(int) int {
	logtrace.LogWithFunctionName()
	return func(count int) int {
		if count > max {
			count = max
		}

		if count <= min {
			return count
		}

		return rand.Intn(count-min) + min
	}
}

func RandomOfTotal() func(count int) int {
	logtrace.LogWithFunctionName()
	return func(count int) int {
		return rand.Intn(count) + 1
	}
}

func Percentage(pct uint8) func(count int) int {
	logtrace.LogWithFunctionName()
	adjustedPct := float64(pct) / 100
	return func(count int) int {
		return int(float64(count) * adjustedPct)
	}
}

func PercentageRange(a uint8, b uint8) func(count int) int {
	logtrace.LogWithFunctionName()
	minVal := min(a, b)
	maxVal := max(a, b)
	delta := maxVal - minVal
	if delta == 0 {
		return Percentage(minVal)
	}
	return func(count int) int {
		pct := minVal + uint8(rand.Int31n(int32(delta)))
		adjustedPct := float64(pct) / 100
		return int(float64(count) * adjustedPct)
	}
}

func SelectRandom(run model.Run, selector string, f func(count int) int) ([]*model.Component, error) {
	logtrace.LogWithFunctionName()
	list := run.GetModel().SelectComponents(selector)
	toSelect := f(len(list))

	if toSelect < 1 {
		return nil, nil
	}

	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})

	var result []*model.Component
	for i := 0; i < len(list) && i < toSelect; i++ {
		result = append(result, list[i])
	}
	return result, nil
}

func StopSelected(run model.Run, list []*model.Component, concurrency int) error {
	logtrace.LogWithFunctionName()
	if len(list) == 0 {
		return nil
	}
	return run.GetModel().ForEachComponentIn(list, concurrency, func(c *model.Component) error {
		if _, ok := c.Type.(model.ServerComponent); ok {
			if err := c.Type.Stop(run, c); err != nil {
				return err
			}

			for {
				isRunning, err := c.IsRunning(run)
				if err != nil {
					return err
				}
				if !isRunning {
					break
				} else {
					time.Sleep(250 * time.Millisecond)
				}
			}
			time.Sleep(time.Second)
			return nil
		}
		return fmt.Errorf("component %v isn't of ServerComponent type, is of type %T", c, c.Type)
	})
}

func RestartSelected(run model.Run, concurrency int, list ...*model.Component) error {
	logtrace.LogWithFunctionName()
	if len(list) == 0 {
		return nil
	}
	return run.GetModel().ForEachComponentIn(list, concurrency, func(c *model.Component) error {
		if sc, ok := c.Type.(model.ServerComponent); ok {
			if err := c.Type.Stop(run, c); err != nil {
				return err
			}

			for {
				isRunning, err := c.IsRunning(run)
				if err != nil {
					return err
				}
				if !isRunning {
					break
				} else {
					time.Sleep(250 * time.Millisecond)
				}
			}
			time.Sleep(time.Second)
			return sc.Start(run, c)
		}
		return fmt.Errorf("component %v isn't of ServerComponent type, is of type %T", c, c.Type)
	})
}

func ValidateUp(run model.Run, spec string, concurrency int, timeout time.Duration) error {
	logtrace.LogWithFunctionName()
	start := time.Now()
	components := run.GetModel().SelectComponents(spec)
	pfxlog.Logger().Infof("checking if all %v components for spec '%s' are running", len(components), spec)
	err := run.GetModel().ForEachComponentIn(components, concurrency, func(c *model.Component) error {
		for {
			isRunning, err := c.IsRunning(run)
			if err != nil {
				return err
			}
			if isRunning {
				return nil
			}
			if time.Since(start) > timeout {
				return fmt.Errorf("timed out waiting for component %s to be running", c.Id)
			}
			time.Sleep(time.Second)
		}
	})
	if err == nil {
		pfxlog.Logger().Infof("all %v components for spec '%s' are running", len(components), spec)
	}
	return err
}

func EnsureLoggedIntoCtrl(run model.Run, c *model.Component, timeout time.Duration) (*zitirest.Clients, error) {
	logtrace.LogWithFunctionName()
	username := c.MustStringVariable("credentials.edge.username")
	password := c.MustStringVariable("credentials.edge.password")
	edgeApiBaseUrl := c.Host.PublicIp + ":1280"

	var clients *zitirest.Clients
	loginStart := time.Now()
	for {
		var err error
		clients, err = zitirest.NewManagementClients(edgeApiBaseUrl)
		if err != nil {
			if time.Since(loginStart) > timeout {
				return nil, err
			}
			pfxlog.Logger().WithField("ctrlId", c.Id).WithError(err).Info("failed to initialize mgmt client, trying again in 1s")
			if err = EnsureRunning(c, run); err != nil {
				pfxlog.Logger().WithField("ctrlId", c.Id).WithError(err).Info("error while trying to ensure ctrl running")
			}
			time.Sleep(time.Second)
			continue
		}

		if err = clients.Authenticate(username, password); err != nil {
			if time.Since(loginStart) > timeout {
				return nil, err
			}
			pfxlog.Logger().WithField("ctrlId", c.Id).WithError(err).Info("failed to authenticate, trying again in 1s")
			if err = EnsureRunning(c, run); err != nil {
				pfxlog.Logger().WithField("ctrlId", c.Id).WithError(err).Info("error while trying to ensure ctrl running")
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return clients, nil
}

func EnsureRunning(c *model.Component, run model.Run) error {
	logtrace.LogWithFunctionName()
	if sc, ok := c.Type.(model.ServerComponent); ok {
		isRunning, err := c.IsRunning(run)
		if err != nil {
			return err
		}
		if isRunning {
			return nil
		}
		time.Sleep(time.Second)
		return sc.Start(run, c)
	}
	return fmt.Errorf("component %v isn't of ServerComponent type, is of type %T", c, c.Type)
}

func Randomize[T any](s []T) {
	logtrace.LogWithFunctionName()
	for i := 0; i < len(s); i++ {
		idx := rand.Intn(len(s))
		e1 := s[i]
		e2 := s[idx]
		s[i] = e2
		s[idx] = e1
	}
}
