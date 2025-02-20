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

package link

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

// NewDeleteLinkParams creates a new DeleteLinkParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewDeleteLinkParams() *DeleteLinkParams {
    logtrace.LogWithFunctionName()
	return &DeleteLinkParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewDeleteLinkParamsWithTimeout creates a new DeleteLinkParams object
// with the ability to set a timeout on a request.
func NewDeleteLinkParamsWithTimeout(timeout time.Duration) *DeleteLinkParams {
    logtrace.LogWithFunctionName()
	return &DeleteLinkParams{
		timeout: timeout,
	}
}

// NewDeleteLinkParamsWithContext creates a new DeleteLinkParams object
// with the ability to set a context for a request.
func NewDeleteLinkParamsWithContext(ctx context.Context) *DeleteLinkParams {
    logtrace.LogWithFunctionName()
	return &DeleteLinkParams{
		Context: ctx,
	}
}

// NewDeleteLinkParamsWithHTTPClient creates a new DeleteLinkParams object
// with the ability to set a custom HTTPClient for a request.
func NewDeleteLinkParamsWithHTTPClient(client *http.Client) *DeleteLinkParams {
    logtrace.LogWithFunctionName()
	return &DeleteLinkParams{
		HTTPClient: client,
	}
}

/* DeleteLinkParams contains all the parameters to send to the API endpoint
   for the delete link operation.

   Typically these are written to a http.Request.
*/
type DeleteLinkParams struct {

	/* ID.

	   The id of the requested resource
	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the delete link params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeleteLinkParams) WithDefaults() *DeleteLinkParams {
    logtrace.LogWithFunctionName()
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the delete link params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeleteLinkParams) SetDefaults() {
    logtrace.LogWithFunctionName()
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the delete link params
func (o *DeleteLinkParams) WithTimeout(timeout time.Duration) *DeleteLinkParams {
    logtrace.LogWithFunctionName()
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete link params
func (o *DeleteLinkParams) SetTimeout(timeout time.Duration) {
    logtrace.LogWithFunctionName()
	o.timeout = timeout
}

// WithContext adds the context to the delete link params
func (o *DeleteLinkParams) WithContext(ctx context.Context) *DeleteLinkParams {
    logtrace.LogWithFunctionName()
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete link params
func (o *DeleteLinkParams) SetContext(ctx context.Context) {
    logtrace.LogWithFunctionName()
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete link params
func (o *DeleteLinkParams) WithHTTPClient(client *http.Client) *DeleteLinkParams {
    logtrace.LogWithFunctionName()
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete link params
func (o *DeleteLinkParams) SetHTTPClient(client *http.Client) {
    logtrace.LogWithFunctionName()
	o.HTTPClient = client
}

// WithID adds the id to the delete link params
func (o *DeleteLinkParams) WithID(id string) *DeleteLinkParams {
    logtrace.LogWithFunctionName()
	o.SetID(id)
	return o
}

// SetID adds the id to the delete link params
func (o *DeleteLinkParams) SetID(id string) {
    logtrace.LogWithFunctionName()
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *DeleteLinkParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param id
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
