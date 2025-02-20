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

package database

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
)

// NewCheckDataIntegrityParams creates a new CheckDataIntegrityParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewCheckDataIntegrityParams() *CheckDataIntegrityParams {
    logtrace.LogWithFunctionName()
	return &CheckDataIntegrityParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewCheckDataIntegrityParamsWithTimeout creates a new CheckDataIntegrityParams object
// with the ability to set a timeout on a request.
func NewCheckDataIntegrityParamsWithTimeout(timeout time.Duration) *CheckDataIntegrityParams {
    logtrace.LogWithFunctionName()
	return &CheckDataIntegrityParams{
		timeout: timeout,
	}
}

// NewCheckDataIntegrityParamsWithContext creates a new CheckDataIntegrityParams object
// with the ability to set a context for a request.
func NewCheckDataIntegrityParamsWithContext(ctx context.Context) *CheckDataIntegrityParams {
    logtrace.LogWithFunctionName()
	return &CheckDataIntegrityParams{
		Context: ctx,
	}
}

// NewCheckDataIntegrityParamsWithHTTPClient creates a new CheckDataIntegrityParams object
// with the ability to set a custom HTTPClient for a request.
func NewCheckDataIntegrityParamsWithHTTPClient(client *http.Client) *CheckDataIntegrityParams {
    logtrace.LogWithFunctionName()
	return &CheckDataIntegrityParams{
		HTTPClient: client,
	}
}

/* CheckDataIntegrityParams contains all the parameters to send to the API endpoint
   for the check data integrity operation.

   Typically these are written to a http.Request.
*/
type CheckDataIntegrityParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the check data integrity params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CheckDataIntegrityParams) WithDefaults() *CheckDataIntegrityParams {
    logtrace.LogWithFunctionName()
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the check data integrity params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CheckDataIntegrityParams) SetDefaults() {
    logtrace.LogWithFunctionName()
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the check data integrity params
func (o *CheckDataIntegrityParams) WithTimeout(timeout time.Duration) *CheckDataIntegrityParams {
    logtrace.LogWithFunctionName()
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the check data integrity params
func (o *CheckDataIntegrityParams) SetTimeout(timeout time.Duration) {
    logtrace.LogWithFunctionName()
	o.timeout = timeout
}

// WithContext adds the context to the check data integrity params
func (o *CheckDataIntegrityParams) WithContext(ctx context.Context) *CheckDataIntegrityParams {
    logtrace.LogWithFunctionName()
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the check data integrity params
func (o *CheckDataIntegrityParams) SetContext(ctx context.Context) {
    logtrace.LogWithFunctionName()
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the check data integrity params
func (o *CheckDataIntegrityParams) WithHTTPClient(client *http.Client) *CheckDataIntegrityParams {
    logtrace.LogWithFunctionName()
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the check data integrity params
func (o *CheckDataIntegrityParams) SetHTTPClient(client *http.Client) {
    logtrace.LogWithFunctionName()
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *CheckDataIntegrityParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
