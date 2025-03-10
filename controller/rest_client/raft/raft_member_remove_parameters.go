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

package raft

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"ztna-core/ztna/controller/rest_model"
)

// NewRaftMemberRemoveParams creates a new RaftMemberRemoveParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewRaftMemberRemoveParams() *RaftMemberRemoveParams {
	return &RaftMemberRemoveParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewRaftMemberRemoveParamsWithTimeout creates a new RaftMemberRemoveParams object
// with the ability to set a timeout on a request.
func NewRaftMemberRemoveParamsWithTimeout(timeout time.Duration) *RaftMemberRemoveParams {
	return &RaftMemberRemoveParams{
		timeout: timeout,
	}
}

// NewRaftMemberRemoveParamsWithContext creates a new RaftMemberRemoveParams object
// with the ability to set a context for a request.
func NewRaftMemberRemoveParamsWithContext(ctx context.Context) *RaftMemberRemoveParams {
	return &RaftMemberRemoveParams{
		Context: ctx,
	}
}

// NewRaftMemberRemoveParamsWithHTTPClient creates a new RaftMemberRemoveParams object
// with the ability to set a custom HTTPClient for a request.
func NewRaftMemberRemoveParamsWithHTTPClient(client *http.Client) *RaftMemberRemoveParams {
	return &RaftMemberRemoveParams{
		HTTPClient: client,
	}
}

/* RaftMemberRemoveParams contains all the parameters to send to the API endpoint
   for the raft member remove operation.

   Typically these are written to a http.Request.
*/
type RaftMemberRemoveParams struct {

	/* Member.

	   member parameters
	*/
	Member *rest_model.RaftMemberRemove

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the raft member remove params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *RaftMemberRemoveParams) WithDefaults() *RaftMemberRemoveParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the raft member remove params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *RaftMemberRemoveParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the raft member remove params
func (o *RaftMemberRemoveParams) WithTimeout(timeout time.Duration) *RaftMemberRemoveParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the raft member remove params
func (o *RaftMemberRemoveParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the raft member remove params
func (o *RaftMemberRemoveParams) WithContext(ctx context.Context) *RaftMemberRemoveParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the raft member remove params
func (o *RaftMemberRemoveParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the raft member remove params
func (o *RaftMemberRemoveParams) WithHTTPClient(client *http.Client) *RaftMemberRemoveParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the raft member remove params
func (o *RaftMemberRemoveParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithMember adds the member to the raft member remove params
func (o *RaftMemberRemoveParams) WithMember(member *rest_model.RaftMemberRemove) *RaftMemberRemoveParams {
	o.SetMember(member)
	return o
}

// SetMember adds the member to the raft member remove params
func (o *RaftMemberRemoveParams) SetMember(member *rest_model.RaftMemberRemove) {
	o.Member = member
}

// WriteToRequest writes these params to a swagger request
func (o *RaftMemberRemoveParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Member != nil {
		if err := r.SetBodyParam(o.Member); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
