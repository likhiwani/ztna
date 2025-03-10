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
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"ztna-core/ztna/controller/rest_model"
)

// DetailLinkReader is a Reader for the DetailLink structure.
type DetailLinkReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DetailLinkReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewDetailLinkOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewDetailLinkUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewDetailLinkNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 429:
		result := NewDetailLinkTooManyRequests()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewDetailLinkOK creates a DetailLinkOK with default headers values
func NewDetailLinkOK() *DetailLinkOK {
	return &DetailLinkOK{}
}

/* DetailLinkOK describes a response with status code 200, with default header values.

A single link
*/
type DetailLinkOK struct {
	Payload *rest_model.DetailLinkEnvelope
}

func (o *DetailLinkOK) Error() string {
	return fmt.Sprintf("[GET /links/{id}][%d] detailLinkOK  %+v", 200, o.Payload)
}
func (o *DetailLinkOK) GetPayload() *rest_model.DetailLinkEnvelope {
	return o.Payload
}

func (o *DetailLinkOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.DetailLinkEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDetailLinkUnauthorized creates a DetailLinkUnauthorized with default headers values
func NewDetailLinkUnauthorized() *DetailLinkUnauthorized {
	return &DetailLinkUnauthorized{}
}

/* DetailLinkUnauthorized describes a response with status code 401, with default header values.

The currently supplied session does not have the correct access rights to request this resource
*/
type DetailLinkUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *DetailLinkUnauthorized) Error() string {
	return fmt.Sprintf("[GET /links/{id}][%d] detailLinkUnauthorized  %+v", 401, o.Payload)
}
func (o *DetailLinkUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *DetailLinkUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDetailLinkNotFound creates a DetailLinkNotFound with default headers values
func NewDetailLinkNotFound() *DetailLinkNotFound {
	return &DetailLinkNotFound{}
}

/* DetailLinkNotFound describes a response with status code 404, with default header values.

The requested resource does not exist
*/
type DetailLinkNotFound struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *DetailLinkNotFound) Error() string {
	return fmt.Sprintf("[GET /links/{id}][%d] detailLinkNotFound  %+v", 404, o.Payload)
}
func (o *DetailLinkNotFound) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *DetailLinkNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDetailLinkTooManyRequests creates a DetailLinkTooManyRequests with default headers values
func NewDetailLinkTooManyRequests() *DetailLinkTooManyRequests {
	return &DetailLinkTooManyRequests{}
}

/* DetailLinkTooManyRequests describes a response with status code 429, with default header values.

The resource requested is rate limited and the rate limit has been exceeded
*/
type DetailLinkTooManyRequests struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *DetailLinkTooManyRequests) Error() string {
	return fmt.Sprintf("[GET /links/{id}][%d] detailLinkTooManyRequests  %+v", 429, o.Payload)
}
func (o *DetailLinkTooManyRequests) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *DetailLinkTooManyRequests) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
