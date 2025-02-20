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

package circuit

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

// ListCircuitsReader is a Reader for the ListCircuits structure.
type ListCircuitsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListCircuitsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
    logtrace.LogWithFunctionName()
	switch response.Code() {
	case 200:
		result := NewListCircuitsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewListCircuitsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 429:
		result := NewListCircuitsTooManyRequests()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListCircuitsOK creates a ListCircuitsOK with default headers values
func NewListCircuitsOK() *ListCircuitsOK {
    logtrace.LogWithFunctionName()
	return &ListCircuitsOK{}
}

/* ListCircuitsOK describes a response with status code 200, with default header values.

A list of circuits
*/
type ListCircuitsOK struct {
	Payload *rest_model.ListCircuitsEnvelope
}

func (o *ListCircuitsOK) Error() string {
    logtrace.LogWithFunctionName()
	return fmt.Sprintf("[GET /circuits][%d] listCircuitsOK  %+v", 200, o.Payload)
}
func (o *ListCircuitsOK) GetPayload() *rest_model.ListCircuitsEnvelope {
    logtrace.LogWithFunctionName()
	return o.Payload
}

func (o *ListCircuitsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	o.Payload = new(rest_model.ListCircuitsEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListCircuitsUnauthorized creates a ListCircuitsUnauthorized with default headers values
func NewListCircuitsUnauthorized() *ListCircuitsUnauthorized {
    logtrace.LogWithFunctionName()
	return &ListCircuitsUnauthorized{}
}

/* ListCircuitsUnauthorized describes a response with status code 401, with default header values.

The currently supplied session does not have the correct access rights to request this resource
*/
type ListCircuitsUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *ListCircuitsUnauthorized) Error() string {
    logtrace.LogWithFunctionName()
	return fmt.Sprintf("[GET /circuits][%d] listCircuitsUnauthorized  %+v", 401, o.Payload)
}
func (o *ListCircuitsUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
    logtrace.LogWithFunctionName()
	return o.Payload
}

func (o *ListCircuitsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListCircuitsTooManyRequests creates a ListCircuitsTooManyRequests with default headers values
func NewListCircuitsTooManyRequests() *ListCircuitsTooManyRequests {
    logtrace.LogWithFunctionName()
	return &ListCircuitsTooManyRequests{}
}

/* ListCircuitsTooManyRequests describes a response with status code 429, with default header values.

The resource requested is rate limited and the rate limit has been exceeded
*/
type ListCircuitsTooManyRequests struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *ListCircuitsTooManyRequests) Error() string {
    logtrace.LogWithFunctionName()
	return fmt.Sprintf("[GET /circuits][%d] listCircuitsTooManyRequests  %+v", 429, o.Payload)
}
func (o *ListCircuitsTooManyRequests) GetPayload() *rest_model.APIErrorEnvelope {
    logtrace.LogWithFunctionName()
	return o.Payload
}

func (o *ListCircuitsTooManyRequests) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
