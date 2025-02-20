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

package service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"ztna-core/ztna/logtrace"
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"ztna-core/ztna/controller/rest_model"
)

// ListServiceTerminatorsReader is a Reader for the ListServiceTerminators structure.
type ListServiceTerminatorsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListServiceTerminatorsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
    logtrace.LogWithFunctionName()
	switch response.Code() {
	case 200:
		result := NewListServiceTerminatorsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewListServiceTerminatorsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewListServiceTerminatorsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 429:
		result := NewListServiceTerminatorsTooManyRequests()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListServiceTerminatorsOK creates a ListServiceTerminatorsOK with default headers values
func NewListServiceTerminatorsOK() *ListServiceTerminatorsOK {
    logtrace.LogWithFunctionName()
	return &ListServiceTerminatorsOK{}
}

/* ListServiceTerminatorsOK describes a response with status code 200, with default header values.

A list of terminators
*/
type ListServiceTerminatorsOK struct {
	Payload *rest_model.ListTerminatorsEnvelope
}

func (o *ListServiceTerminatorsOK) Error() string {
    logtrace.LogWithFunctionName()
	return fmt.Sprintf("[GET /services/{id}/terminators][%d] listServiceTerminatorsOK  %+v", 200, o.Payload)
}
func (o *ListServiceTerminatorsOK) GetPayload() *rest_model.ListTerminatorsEnvelope {
    logtrace.LogWithFunctionName()
	return o.Payload
}

func (o *ListServiceTerminatorsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	o.Payload = new(rest_model.ListTerminatorsEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListServiceTerminatorsBadRequest creates a ListServiceTerminatorsBadRequest with default headers values
func NewListServiceTerminatorsBadRequest() *ListServiceTerminatorsBadRequest {
    logtrace.LogWithFunctionName()
	return &ListServiceTerminatorsBadRequest{}
}

/* ListServiceTerminatorsBadRequest describes a response with status code 400, with default header values.

The supplied request contains invalid fields or could not be parsed (json and non-json bodies). The error's code, message, and cause fields can be inspected for further information
*/
type ListServiceTerminatorsBadRequest struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *ListServiceTerminatorsBadRequest) Error() string {
    logtrace.LogWithFunctionName()
	return fmt.Sprintf("[GET /services/{id}/terminators][%d] listServiceTerminatorsBadRequest  %+v", 400, o.Payload)
}
func (o *ListServiceTerminatorsBadRequest) GetPayload() *rest_model.APIErrorEnvelope {
    logtrace.LogWithFunctionName()
	return o.Payload
}

func (o *ListServiceTerminatorsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListServiceTerminatorsUnauthorized creates a ListServiceTerminatorsUnauthorized with default headers values
func NewListServiceTerminatorsUnauthorized() *ListServiceTerminatorsUnauthorized {
    logtrace.LogWithFunctionName()
	return &ListServiceTerminatorsUnauthorized{}
}

/* ListServiceTerminatorsUnauthorized describes a response with status code 401, with default header values.

The currently supplied session does not have the correct access rights to request this resource
*/
type ListServiceTerminatorsUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *ListServiceTerminatorsUnauthorized) Error() string {
    logtrace.LogWithFunctionName()
	return fmt.Sprintf("[GET /services/{id}/terminators][%d] listServiceTerminatorsUnauthorized  %+v", 401, o.Payload)
}
func (o *ListServiceTerminatorsUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
    logtrace.LogWithFunctionName()
	return o.Payload
}

func (o *ListServiceTerminatorsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListServiceTerminatorsTooManyRequests creates a ListServiceTerminatorsTooManyRequests with default headers values
func NewListServiceTerminatorsTooManyRequests() *ListServiceTerminatorsTooManyRequests {
    logtrace.LogWithFunctionName()
	return &ListServiceTerminatorsTooManyRequests{}
}

/* ListServiceTerminatorsTooManyRequests describes a response with status code 429, with default header values.

The resource requested is rate limited and the rate limit has been exceeded
*/
type ListServiceTerminatorsTooManyRequests struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *ListServiceTerminatorsTooManyRequests) Error() string {
    logtrace.LogWithFunctionName()
	return fmt.Sprintf("[GET /services/{id}/terminators][%d] listServiceTerminatorsTooManyRequests  %+v", 429, o.Payload)
}
func (o *ListServiceTerminatorsTooManyRequests) GetPayload() *rest_model.APIErrorEnvelope {
    logtrace.LogWithFunctionName()
	return o.Payload
}

func (o *ListServiceTerminatorsTooManyRequests) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
