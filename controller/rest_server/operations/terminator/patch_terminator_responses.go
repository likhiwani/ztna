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
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"ztna-core/ztna/logtrace"
	"net/http"

	"github.com/go-openapi/runtime"

	"ztna-core/ztna/controller/rest_model"
)

// PatchTerminatorOKCode is the HTTP code returned for type PatchTerminatorOK
const PatchTerminatorOKCode int = 200

/*PatchTerminatorOK The patch request was successful and the resource has been altered

swagger:response patchTerminatorOK
*/
type PatchTerminatorOK struct {

	/*
	  In: Body
	*/
	Payload *rest_model.Empty `json:"body,omitempty"`
}

// NewPatchTerminatorOK creates PatchTerminatorOK with default headers values
func NewPatchTerminatorOK() *PatchTerminatorOK {
    logtrace.LogWithFunctionName()

	return &PatchTerminatorOK{}
}

// WithPayload adds the payload to the patch terminator o k response
func (o *PatchTerminatorOK) WithPayload(payload *rest_model.Empty) *PatchTerminatorOK {
    logtrace.LogWithFunctionName()
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the patch terminator o k response
func (o *PatchTerminatorOK) SetPayload(payload *rest_model.Empty) {
    logtrace.LogWithFunctionName()
	o.Payload = payload
}

// WriteResponse to the client
func (o *PatchTerminatorOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
    logtrace.LogWithFunctionName()

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PatchTerminatorBadRequestCode is the HTTP code returned for type PatchTerminatorBadRequest
const PatchTerminatorBadRequestCode int = 400

/*PatchTerminatorBadRequest The supplied request contains invalid fields or could not be parsed (json and non-json bodies). The error's code, message, and cause fields can be inspected for further information

swagger:response patchTerminatorBadRequest
*/
type PatchTerminatorBadRequest struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewPatchTerminatorBadRequest creates PatchTerminatorBadRequest with default headers values
func NewPatchTerminatorBadRequest() *PatchTerminatorBadRequest {
    logtrace.LogWithFunctionName()

	return &PatchTerminatorBadRequest{}
}

// WithPayload adds the payload to the patch terminator bad request response
func (o *PatchTerminatorBadRequest) WithPayload(payload *rest_model.APIErrorEnvelope) *PatchTerminatorBadRequest {
    logtrace.LogWithFunctionName()
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the patch terminator bad request response
func (o *PatchTerminatorBadRequest) SetPayload(payload *rest_model.APIErrorEnvelope) {
    logtrace.LogWithFunctionName()
	o.Payload = payload
}

// WriteResponse to the client
func (o *PatchTerminatorBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
    logtrace.LogWithFunctionName()

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PatchTerminatorUnauthorizedCode is the HTTP code returned for type PatchTerminatorUnauthorized
const PatchTerminatorUnauthorizedCode int = 401

/*PatchTerminatorUnauthorized The currently supplied session does not have the correct access rights to request this resource

swagger:response patchTerminatorUnauthorized
*/
type PatchTerminatorUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewPatchTerminatorUnauthorized creates PatchTerminatorUnauthorized with default headers values
func NewPatchTerminatorUnauthorized() *PatchTerminatorUnauthorized {
    logtrace.LogWithFunctionName()

	return &PatchTerminatorUnauthorized{}
}

// WithPayload adds the payload to the patch terminator unauthorized response
func (o *PatchTerminatorUnauthorized) WithPayload(payload *rest_model.APIErrorEnvelope) *PatchTerminatorUnauthorized {
    logtrace.LogWithFunctionName()
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the patch terminator unauthorized response
func (o *PatchTerminatorUnauthorized) SetPayload(payload *rest_model.APIErrorEnvelope) {
    logtrace.LogWithFunctionName()
	o.Payload = payload
}

// WriteResponse to the client
func (o *PatchTerminatorUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
    logtrace.LogWithFunctionName()

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PatchTerminatorNotFoundCode is the HTTP code returned for type PatchTerminatorNotFound
const PatchTerminatorNotFoundCode int = 404

/*PatchTerminatorNotFound The requested resource does not exist

swagger:response patchTerminatorNotFound
*/
type PatchTerminatorNotFound struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewPatchTerminatorNotFound creates PatchTerminatorNotFound with default headers values
func NewPatchTerminatorNotFound() *PatchTerminatorNotFound {
    logtrace.LogWithFunctionName()

	return &PatchTerminatorNotFound{}
}

// WithPayload adds the payload to the patch terminator not found response
func (o *PatchTerminatorNotFound) WithPayload(payload *rest_model.APIErrorEnvelope) *PatchTerminatorNotFound {
    logtrace.LogWithFunctionName()
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the patch terminator not found response
func (o *PatchTerminatorNotFound) SetPayload(payload *rest_model.APIErrorEnvelope) {
    logtrace.LogWithFunctionName()
	o.Payload = payload
}

// WriteResponse to the client
func (o *PatchTerminatorNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
    logtrace.LogWithFunctionName()

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PatchTerminatorTooManyRequestsCode is the HTTP code returned for type PatchTerminatorTooManyRequests
const PatchTerminatorTooManyRequestsCode int = 429

/*PatchTerminatorTooManyRequests The resource requested is rate limited and the rate limit has been exceeded

swagger:response patchTerminatorTooManyRequests
*/
type PatchTerminatorTooManyRequests struct {

	/*
	  In: Body
	*/
	Payload *rest_model.APIErrorEnvelope `json:"body,omitempty"`
}

// NewPatchTerminatorTooManyRequests creates PatchTerminatorTooManyRequests with default headers values
func NewPatchTerminatorTooManyRequests() *PatchTerminatorTooManyRequests {
    logtrace.LogWithFunctionName()

	return &PatchTerminatorTooManyRequests{}
}

// WithPayload adds the payload to the patch terminator too many requests response
func (o *PatchTerminatorTooManyRequests) WithPayload(payload *rest_model.APIErrorEnvelope) *PatchTerminatorTooManyRequests {
    logtrace.LogWithFunctionName()
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the patch terminator too many requests response
func (o *PatchTerminatorTooManyRequests) SetPayload(payload *rest_model.APIErrorEnvelope) {
    logtrace.LogWithFunctionName()
	o.Payload = payload
}

// WriteResponse to the client
func (o *PatchTerminatorTooManyRequests) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
    logtrace.LogWithFunctionName()

	rw.WriteHeader(429)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
