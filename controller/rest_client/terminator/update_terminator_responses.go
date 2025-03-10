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
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"ztna-core/ztna/controller/rest_model"
)

// UpdateTerminatorReader is a Reader for the UpdateTerminator structure.
type UpdateTerminatorReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UpdateTerminatorReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUpdateTerminatorOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewUpdateTerminatorBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewUpdateTerminatorUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewUpdateTerminatorNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 429:
		result := NewUpdateTerminatorTooManyRequests()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewUpdateTerminatorOK creates a UpdateTerminatorOK with default headers values
func NewUpdateTerminatorOK() *UpdateTerminatorOK {
	return &UpdateTerminatorOK{}
}

/* UpdateTerminatorOK describes a response with status code 200, with default header values.

The update request was successful and the resource has been altered
*/
type UpdateTerminatorOK struct {
	Payload *rest_model.Empty
}

func (o *UpdateTerminatorOK) Error() string {
	return fmt.Sprintf("[PUT /terminators/{id}][%d] updateTerminatorOK  %+v", 200, o.Payload)
}
func (o *UpdateTerminatorOK) GetPayload() *rest_model.Empty {
	return o.Payload
}

func (o *UpdateTerminatorOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.Empty)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateTerminatorBadRequest creates a UpdateTerminatorBadRequest with default headers values
func NewUpdateTerminatorBadRequest() *UpdateTerminatorBadRequest {
	return &UpdateTerminatorBadRequest{}
}

/* UpdateTerminatorBadRequest describes a response with status code 400, with default header values.

The supplied request contains invalid fields or could not be parsed (json and non-json bodies). The error's code, message, and cause fields can be inspected for further information
*/
type UpdateTerminatorBadRequest struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *UpdateTerminatorBadRequest) Error() string {
	return fmt.Sprintf("[PUT /terminators/{id}][%d] updateTerminatorBadRequest  %+v", 400, o.Payload)
}
func (o *UpdateTerminatorBadRequest) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *UpdateTerminatorBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateTerminatorUnauthorized creates a UpdateTerminatorUnauthorized with default headers values
func NewUpdateTerminatorUnauthorized() *UpdateTerminatorUnauthorized {
	return &UpdateTerminatorUnauthorized{}
}

/* UpdateTerminatorUnauthorized describes a response with status code 401, with default header values.

The currently supplied session does not have the correct access rights to request this resource
*/
type UpdateTerminatorUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *UpdateTerminatorUnauthorized) Error() string {
	return fmt.Sprintf("[PUT /terminators/{id}][%d] updateTerminatorUnauthorized  %+v", 401, o.Payload)
}
func (o *UpdateTerminatorUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *UpdateTerminatorUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateTerminatorNotFound creates a UpdateTerminatorNotFound with default headers values
func NewUpdateTerminatorNotFound() *UpdateTerminatorNotFound {
	return &UpdateTerminatorNotFound{}
}

/* UpdateTerminatorNotFound describes a response with status code 404, with default header values.

The requested resource does not exist
*/
type UpdateTerminatorNotFound struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *UpdateTerminatorNotFound) Error() string {
	return fmt.Sprintf("[PUT /terminators/{id}][%d] updateTerminatorNotFound  %+v", 404, o.Payload)
}
func (o *UpdateTerminatorNotFound) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *UpdateTerminatorNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateTerminatorTooManyRequests creates a UpdateTerminatorTooManyRequests with default headers values
func NewUpdateTerminatorTooManyRequests() *UpdateTerminatorTooManyRequests {
	return &UpdateTerminatorTooManyRequests{}
}

/* UpdateTerminatorTooManyRequests describes a response with status code 429, with default header values.

The resource requested is rate limited and the rate limit has been exceeded
*/
type UpdateTerminatorTooManyRequests struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *UpdateTerminatorTooManyRequests) Error() string {
	return fmt.Sprintf("[PUT /terminators/{id}][%d] updateTerminatorTooManyRequests  %+v", 429, o.Payload)
}
func (o *UpdateTerminatorTooManyRequests) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *UpdateTerminatorTooManyRequests) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
