// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package terminator

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"ztna-core/ztna/logtrace"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// UpdateTerminatorHandlerFunc turns a function with the right signature into a update terminator handler
type UpdateTerminatorHandlerFunc func(UpdateTerminatorParams) middleware.Responder

// Handle executing the request and returning a response
func (fn UpdateTerminatorHandlerFunc) Handle(params UpdateTerminatorParams) middleware.Responder {
    logtrace.LogWithFunctionName()
	return fn(params)
}

// UpdateTerminatorHandler interface for that can handle valid update terminator params
type UpdateTerminatorHandler interface {
	Handle(UpdateTerminatorParams) middleware.Responder
}

// NewUpdateTerminator creates a new http.Handler for the update terminator operation
func NewUpdateTerminator(ctx *middleware.Context, handler UpdateTerminatorHandler) *UpdateTerminator {
    logtrace.LogWithFunctionName()
	return &UpdateTerminator{Context: ctx, Handler: handler}
}

/* UpdateTerminator swagger:route PUT /terminators/{id} Terminator updateTerminator

Update all fields on a terminator

Update all fields on a terminator by id. Requires admin access.

*/
type UpdateTerminator struct {
	Context *middleware.Context
	Handler UpdateTerminatorHandler
}

func (o *UpdateTerminator) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
    logtrace.LogWithFunctionName()
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewUpdateTerminatorParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
