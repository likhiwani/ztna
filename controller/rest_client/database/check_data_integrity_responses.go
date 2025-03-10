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
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"ztna-core/ztna/controller/rest_model"
)

// CheckDataIntegrityReader is a Reader for the CheckDataIntegrity structure.
type CheckDataIntegrityReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CheckDataIntegrityReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 202:
		result := NewCheckDataIntegrityAccepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewCheckDataIntegrityUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 429:
		result := NewCheckDataIntegrityTooManyRequests()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewCheckDataIntegrityAccepted creates a CheckDataIntegrityAccepted with default headers values
func NewCheckDataIntegrityAccepted() *CheckDataIntegrityAccepted {
	return &CheckDataIntegrityAccepted{}
}

/* CheckDataIntegrityAccepted describes a response with status code 202, with default header values.

Base empty response
*/
type CheckDataIntegrityAccepted struct {
	Payload *rest_model.Empty
}

func (o *CheckDataIntegrityAccepted) Error() string {
	return fmt.Sprintf("[POST /database/check-data-integrity][%d] checkDataIntegrityAccepted  %+v", 202, o.Payload)
}
func (o *CheckDataIntegrityAccepted) GetPayload() *rest_model.Empty {
	return o.Payload
}

func (o *CheckDataIntegrityAccepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.Empty)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCheckDataIntegrityUnauthorized creates a CheckDataIntegrityUnauthorized with default headers values
func NewCheckDataIntegrityUnauthorized() *CheckDataIntegrityUnauthorized {
	return &CheckDataIntegrityUnauthorized{}
}

/* CheckDataIntegrityUnauthorized describes a response with status code 401, with default header values.

The currently supplied session does not have the correct access rights to request this resource
*/
type CheckDataIntegrityUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *CheckDataIntegrityUnauthorized) Error() string {
	return fmt.Sprintf("[POST /database/check-data-integrity][%d] checkDataIntegrityUnauthorized  %+v", 401, o.Payload)
}
func (o *CheckDataIntegrityUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *CheckDataIntegrityUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCheckDataIntegrityTooManyRequests creates a CheckDataIntegrityTooManyRequests with default headers values
func NewCheckDataIntegrityTooManyRequests() *CheckDataIntegrityTooManyRequests {
	return &CheckDataIntegrityTooManyRequests{}
}

/* CheckDataIntegrityTooManyRequests describes a response with status code 429, with default header values.

The resource requested is rate limited and the rate limit has been exceeded
*/
type CheckDataIntegrityTooManyRequests struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *CheckDataIntegrityTooManyRequests) Error() string {
	return fmt.Sprintf("[POST /database/check-data-integrity][%d] checkDataIntegrityTooManyRequests  %+v", 429, o.Payload)
}
func (o *CheckDataIntegrityTooManyRequests) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *CheckDataIntegrityTooManyRequests) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
