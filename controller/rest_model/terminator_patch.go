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
)

// TerminatorPatch terminator patch
//
// swagger:model terminatorPatch
type TerminatorPatch struct {

	// address
	Address string `json:"address,omitempty"`

	// binding
	Binding string `json:"binding,omitempty"`

	// cost
	Cost *TerminatorCost `json:"cost,omitempty"`

	// host Id
	HostID string `json:"hostId,omitempty"`

	// precedence
	Precedence TerminatorPrecedence `json:"precedence,omitempty"`

	// router
	Router string `json:"router,omitempty"`

	// service
	Service string `json:"service,omitempty"`

	// tags
	Tags *Tags `json:"tags,omitempty"`
}

// Validate validates this terminator patch
func (m *TerminatorPatch) Validate(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()
	var res []error

	if err := m.validateCost(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePrecedence(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTags(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *TerminatorPatch) validateCost(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()
	if swag.IsZero(m.Cost) { // not required
		return nil
	}

	if m.Cost != nil {
		if err := m.Cost.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("cost")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("cost")
			}
			return err
		}
	}

	return nil
}

func (m *TerminatorPatch) validatePrecedence(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()
	if swag.IsZero(m.Precedence) { // not required
		return nil
	}

	if err := m.Precedence.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("precedence")
		} else if ce, ok := err.(*errors.CompositeError); ok {
			return ce.ValidateName("precedence")
		}
		return err
	}

	return nil
}

func (m *TerminatorPatch) validateTags(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()
	if swag.IsZero(m.Tags) { // not required
		return nil
	}

	if m.Tags != nil {
		if err := m.Tags.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("tags")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("tags")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this terminator patch based on the context it is used
func (m *TerminatorPatch) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()
	var res []error

	if err := m.contextValidateCost(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidatePrecedence(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateTags(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *TerminatorPatch) contextValidateCost(ctx context.Context, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if m.Cost != nil {
		if err := m.Cost.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("cost")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("cost")
			}
			return err
		}
	}

	return nil
}

func (m *TerminatorPatch) contextValidatePrecedence(ctx context.Context, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := m.Precedence.ContextValidate(ctx, formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("precedence")
		} else if ce, ok := err.(*errors.CompositeError); ok {
			return ce.ValidateName("precedence")
		}
		return err
	}

	return nil
}

func (m *TerminatorPatch) contextValidateTags(ctx context.Context, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if m.Tags != nil {
		if err := m.Tags.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("tags")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("tags")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *TerminatorPatch) MarshalBinary() ([]byte, error) {
    logtrace.LogWithFunctionName()
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *TerminatorPatch) UnmarshalBinary(b []byte) error {
    logtrace.LogWithFunctionName()
	var res TerminatorPatch
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
