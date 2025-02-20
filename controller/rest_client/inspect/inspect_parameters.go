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

package inspect

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"ztna-core/ztna/logtrace"
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"ztna-core/ztna/controller/rest_model"
)

// NewInspectParams creates a new InspectParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewInspectParams() *InspectParams {
    logtrace.LogWithFunctionName()
	return &InspectParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewInspectParamsWithTimeout creates a new InspectParams object
// with the ability to set a timeout on a request.
func NewInspectParamsWithTimeout(timeout time.Duration) *InspectParams {
    logtrace.LogWithFunctionName()
	return &InspectParams{
		timeout: timeout,
	}
}

// NewInspectParamsWithContext creates a new InspectParams object
// with the ability to set a context for a request.
func NewInspectParamsWithContext(ctx context.Context) *InspectParams {
    logtrace.LogWithFunctionName()
	return &InspectParams{
		Context: ctx,
	}
}

// NewInspectParamsWithHTTPClient creates a new InspectParams object
// with the ability to set a custom HTTPClient for a request.
func NewInspectParamsWithHTTPClient(client *http.Client) *InspectParams {
    logtrace.LogWithFunctionName()
	return &InspectParams{
		HTTPClient: client,
	}
}

/* InspectParams contains all the parameters to send to the API endpoint
   for the inspect operation.

   Typically these are written to a http.Request.
*/
type InspectParams struct {

	/* Request.

	   An inspect request
	*/
	Request *rest_model.InspectRequest

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the inspect params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *InspectParams) WithDefaults() *InspectParams {
    logtrace.LogWithFunctionName()
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the inspect params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *InspectParams) SetDefaults() {
    logtrace.LogWithFunctionName()
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the inspect params
func (o *InspectParams) WithTimeout(timeout time.Duration) *InspectParams {
    logtrace.LogWithFunctionName()
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the inspect params
func (o *InspectParams) SetTimeout(timeout time.Duration) {
    logtrace.LogWithFunctionName()
	o.timeout = timeout
}

// WithContext adds the context to the inspect params
func (o *InspectParams) WithContext(ctx context.Context) *InspectParams {
    logtrace.LogWithFunctionName()
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the inspect params
func (o *InspectParams) SetContext(ctx context.Context) {
    logtrace.LogWithFunctionName()
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the inspect params
func (o *InspectParams) WithHTTPClient(client *http.Client) *InspectParams {
    logtrace.LogWithFunctionName()
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the inspect params
func (o *InspectParams) SetHTTPClient(client *http.Client) {
    logtrace.LogWithFunctionName()
	o.HTTPClient = client
}

// WithRequest adds the request to the inspect params
func (o *InspectParams) WithRequest(request *rest_model.InspectRequest) *InspectParams {
    logtrace.LogWithFunctionName()
	o.SetRequest(request)
	return o
}

// SetRequest adds the request to the inspect params
func (o *InspectParams) SetRequest(request *rest_model.InspectRequest) {
    logtrace.LogWithFunctionName()
	o.Request = request
}

// WriteToRequest writes these params to a swagger request
func (o *InspectParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Request != nil {
		if err := r.SetBodyParam(o.Request); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
