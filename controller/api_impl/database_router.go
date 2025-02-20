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

package api_impl

import (
	"context"
	"errors"
	"sync/atomic"
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/apierror"
	"ztna-core/ztna/controller/rest_model"
	"ztna-core/ztna/controller/rest_server/operations"
	"ztna-core/ztna/controller/rest_server/operations/database"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/michaelquigley/pfxlog"

	"net/http"
	"sync"
	"time"
	"ztna-core/ztna/controller/network"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewDatabaseRouter()
	AddRouter(r)
}

type integrityCheckOp struct {
	running       atomic.Bool
	results       []*rest_model.DataIntegrityCheckDetail
	lock          sync.Mutex
	fixingErrors  bool
	startTime     *time.Time
	endTime       *time.Time
	err           error
	tooManyErrors bool
}

type DatabaseRouter struct {
	integrityCheck integrityCheckOp
}

func NewDatabaseRouter() *DatabaseRouter {
	logtrace.LogWithFunctionName()
	return &DatabaseRouter{}
}

func (r *DatabaseRouter) Register(fabricApi *operations.ZitiFabricAPI, wrapper RequestWrapper) {
	logtrace.LogWithFunctionName()
	fabricApi.DatabaseCreateDatabaseSnapshotHandler = database.CreateDatabaseSnapshotHandlerFunc(func(params database.CreateDatabaseSnapshotParams, _ interface{}) middleware.Responder {
		return wrapper.WrapRequest(r.CreateSnapshot, params.HTTPRequest, "", "")
	})

	fabricApi.DatabaseCreateDatabaseSnapshotWithPathHandler = database.CreateDatabaseSnapshotWithPathHandlerFunc(func(params database.CreateDatabaseSnapshotWithPathParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) {
			r.CreateSnapshotWithPath(n, rc, params.Snapshot)
		}, params.HTTPRequest, "", "")
	})

	fabricApi.DatabaseCheckDataIntegrityHandler = database.CheckDataIntegrityHandlerFunc(func(params database.CheckDataIntegrityParams, _ interface{}) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.CheckDatastoreIntegrity(n, rc, false) }, params.HTTPRequest, "", "")
	})

	fabricApi.DatabaseFixDataIntegrityHandler = database.FixDataIntegrityHandlerFunc(func(params database.FixDataIntegrityParams, _ interface{}) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.CheckDatastoreIntegrity(n, rc, true) }, params.HTTPRequest, "", "")
	})

	fabricApi.DatabaseDataIntegrityResultsHandler = database.DataIntegrityResultsHandlerFunc(func(params database.DataIntegrityResultsParams, _ interface{}) middleware.Responder {
		return wrapper.WrapRequest(r.GetCheckProgress, params.HTTPRequest, "", "")
	})
}

func (r *DatabaseRouter) CreateSnapshot(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	if err := n.SnapshotDatabase(); err != nil {
		if errors.Is(err, network.DbSnapshotTooFrequentError) {
			rc.RespondWithApiError(apierror.NewRateLimited())
			return
		}
		rc.RespondWithError(err)
		return
	}
	rc.RespondWithEmptyOk()
}

func (r *DatabaseRouter) CreateSnapshotWithPath(n *network.Network, rc api.RequestContext, snapshot *rest_model.DatabaseSnapshotCreate) {
	logtrace.LogWithFunctionName()
	var path string
	if snapshot != nil {
		path = snapshot.Path
	}
	actualPath, err := n.SnapshotDatabaseToFile(path)
	if err != nil {
		if errors.Is(err, network.DbSnapshotTooFrequentError) {
			rc.RespondWithApiError(apierror.NewRateLimited())
			return
		}
		rc.RespondWithError(err)
		return
	}

	result := rest_model.DatabaseSnapshotCreateResultEnvelope{
		Data: &rest_model.DatabaseSnapshotCreateDetails{
			Path: &actualPath,
		},
		Meta: &rest_model.Meta{},
	}

	rc.Respond(result, http.StatusOK)
}

func (r *DatabaseRouter) CheckDatastoreIntegrity(n *network.Network, rc api.RequestContext, fixErrors bool) {
	logtrace.LogWithFunctionName()
	if r.integrityCheck.running.CompareAndSwap(false, true) {
		r.integrityCheck.fixingErrors = fixErrors
		go r.runDataIntegrityCheck(n, rc.NewChangeContext().GetContext(), fixErrors)
		rc.Respond(&rest_model.Empty{Data: map[string]interface{}{}, Meta: &rest_model.Meta{}}, http.StatusAccepted)
	} else {
		rc.RespondWithApiError(apierror.NewRateLimited())
	}
}

func (r *DatabaseRouter) GetCheckProgress(_ *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	integrityCheck := &r.integrityCheck

	integrityCheck.lock.Lock()
	defer integrityCheck.lock.Unlock()

	limit := int64(-1)
	zero := int64(0)
	count := int64(len(integrityCheck.results))

	var err *string
	if integrityCheck.err != nil {
		errStr := integrityCheck.err.Error()
		err = &errStr
	}

	inProgress := integrityCheck.running.Load()

	result := rest_model.DataIntegrityCheckResultEnvelope{
		Data: &rest_model.DataIntegrityCheckDetails{
			EndTime:       (*strfmt.DateTime)(integrityCheck.endTime),
			Error:         err,
			FixingErrors:  &integrityCheck.fixingErrors,
			InProgress:    &inProgress,
			Results:       integrityCheck.results,
			StartTime:     (*strfmt.DateTime)(integrityCheck.startTime),
			TooManyErrors: &integrityCheck.tooManyErrors,
		},
		Meta: &rest_model.Meta{
			Pagination: &rest_model.Pagination{
				Limit:      &limit,
				Offset:     &zero,
				TotalCount: &count,
			},
			FilterableFields: make([]string, 0),
		},
	}

	rc.Respond(result, http.StatusOK)
}

func (r *DatabaseRouter) runDataIntegrityCheck(n *network.Network, ctx context.Context, fixErrors bool) {
	logtrace.LogWithFunctionName()
	defer func() {
		r.integrityCheck.lock.Lock()
		now := time.Now()
		r.integrityCheck.endTime = &now
		r.integrityCheck.running.Store(false)
		r.integrityCheck.lock.Unlock()
	}()

	r.integrityCheck.lock.Lock()
	now := time.Now()
	r.integrityCheck.results = nil
	r.integrityCheck.startTime = &now
	r.integrityCheck.endTime = nil
	r.integrityCheck.err = nil
	r.integrityCheck.tooManyErrors = false
	r.integrityCheck.lock.Unlock()

	logger := pfxlog.Logger()

	errorHandler := func(err error, fixed bool) {
		logger.WithError(err).Warnf("data integrity error reported. fixed? %v", fixed)

		r.integrityCheck.lock.Lock()
		defer r.integrityCheck.lock.Unlock()

		if len(r.integrityCheck.results) < 1000 {
			description := err.Error()
			r.integrityCheck.results = append(r.integrityCheck.results, &rest_model.DataIntegrityCheckDetail{
				Description: &description,
				Fixed:       &fixed,
			})
		} else {
			r.integrityCheck.tooManyErrors = true
		}
	}

	r.integrityCheck.err = n.GetStores().CheckIntegrity(n.GetDb(), ctx, fixErrors, errorHandler)
}
