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

package rest_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"ztna-core/ztna/logtrace"
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// LinkDetail link detail
//
// swagger:model linkDetail
type LinkDetail struct {

	// cost
	// Required: true
	Cost *int64 `json:"cost"`

	// dest latency
	// Required: true
	DestLatency *int64 `json:"destLatency"`

	// dest router
	// Required: true
	DestRouter *EntityRef `json:"destRouter"`

	// down
	// Required: true
	Down *bool `json:"down"`

	// id
	// Required: true
	ID *string `json:"id"`

	// iteration
	// Required: true
	Iteration *int64 `json:"iteration"`

	// protocol
	// Required: true
	Protocol *string `json:"protocol"`

	// source latency
	// Required: true
	SourceLatency *int64 `json:"sourceLatency"`

	// source router
	// Required: true
	SourceRouter *EntityRef `json:"sourceRouter"`

	// state
	// Required: true
	State *string `json:"state"`

	// static cost
	// Required: true
	StaticCost *int64 `json:"staticCost"`
}

// Validate validates this link detail
func (m *LinkDetail) Validate(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()
	var res []error

	if err := m.validateCost(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDestLatency(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDestRouter(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDown(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIteration(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateProtocol(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSourceLatency(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSourceRouter(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateState(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStaticCost(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *LinkDetail) validateCost(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("cost", "body", m.Cost); err != nil {
		return err
	}

	return nil
}

func (m *LinkDetail) validateDestLatency(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("destLatency", "body", m.DestLatency); err != nil {
		return err
	}

	return nil
}

func (m *LinkDetail) validateDestRouter(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("destRouter", "body", m.DestRouter); err != nil {
		return err
	}

	if m.DestRouter != nil {
		if err := m.DestRouter.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("destRouter")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("destRouter")
			}
			return err
		}
	}

	return nil
}

func (m *LinkDetail) validateDown(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("down", "body", m.Down); err != nil {
		return err
	}

	return nil
}

func (m *LinkDetail) validateID(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("id", "body", m.ID); err != nil {
		return err
	}

	return nil
}

func (m *LinkDetail) validateIteration(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("iteration", "body", m.Iteration); err != nil {
		return err
	}

	return nil
}

func (m *LinkDetail) validateProtocol(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("protocol", "body", m.Protocol); err != nil {
		return err
	}

	return nil
}

func (m *LinkDetail) validateSourceLatency(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("sourceLatency", "body", m.SourceLatency); err != nil {
		return err
	}

	return nil
}

func (m *LinkDetail) validateSourceRouter(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("sourceRouter", "body", m.SourceRouter); err != nil {
		return err
	}

	if m.SourceRouter != nil {
		if err := m.SourceRouter.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("sourceRouter")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("sourceRouter")
			}
			return err
		}
	}

	return nil
}

func (m *LinkDetail) validateState(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("state", "body", m.State); err != nil {
		return err
	}

	return nil
}

func (m *LinkDetail) validateStaticCost(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("staticCost", "body", m.StaticCost); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this link detail based on the context it is used
func (m *LinkDetail) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()
	var res []error

	if err := m.contextValidateDestRouter(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateSourceRouter(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *LinkDetail) contextValidateDestRouter(ctx context.Context, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if m.DestRouter != nil {
		if err := m.DestRouter.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("destRouter")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("destRouter")
			}
			return err
		}
	}

	return nil
}

func (m *LinkDetail) contextValidateSourceRouter(ctx context.Context, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if m.SourceRouter != nil {
		if err := m.SourceRouter.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("sourceRouter")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("sourceRouter")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *LinkDetail) MarshalBinary() ([]byte, error) {
    logtrace.LogWithFunctionName()
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *LinkDetail) UnmarshalBinary(b []byte) error {
    logtrace.LogWithFunctionName()
	var res LinkDetail
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
