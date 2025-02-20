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

// RaftMemberListValue raft member list value
//
// swagger:model raftMemberListValue
type RaftMemberListValue struct {

	// address
	// Required: true
	Address *string `json:"address"`

	// connected
	// Required: true
	Connected *bool `json:"connected"`

	// id
	// Required: true
	ID *string `json:"id"`

	// leader
	// Required: true
	Leader *bool `json:"leader"`

	// read only
	// Required: true
	ReadOnly *bool `json:"readOnly"`

	// version
	// Required: true
	Version *string `json:"version"`

	// voter
	// Required: true
	Voter *bool `json:"voter"`
}

// Validate validates this raft member list value
func (m *RaftMemberListValue) Validate(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()
	var res []error

	if err := m.validateAddress(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateConnected(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateLeader(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateReadOnly(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVersion(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVoter(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *RaftMemberListValue) validateAddress(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("address", "body", m.Address); err != nil {
		return err
	}

	return nil
}

func (m *RaftMemberListValue) validateConnected(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("connected", "body", m.Connected); err != nil {
		return err
	}

	return nil
}

func (m *RaftMemberListValue) validateID(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("id", "body", m.ID); err != nil {
		return err
	}

	return nil
}

func (m *RaftMemberListValue) validateLeader(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("leader", "body", m.Leader); err != nil {
		return err
	}

	return nil
}

func (m *RaftMemberListValue) validateReadOnly(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("readOnly", "body", m.ReadOnly); err != nil {
		return err
	}

	return nil
}

func (m *RaftMemberListValue) validateVersion(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("version", "body", m.Version); err != nil {
		return err
	}

	return nil
}

func (m *RaftMemberListValue) validateVoter(formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()

	if err := validate.Required("voter", "body", m.Voter); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this raft member list value based on context it is used
func (m *RaftMemberListValue) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
    logtrace.LogWithFunctionName()
	return nil
}

// MarshalBinary interface implementation
func (m *RaftMemberListValue) MarshalBinary() ([]byte, error) {
    logtrace.LogWithFunctionName()
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RaftMemberListValue) UnmarshalBinary(b []byte) error {
    logtrace.LogWithFunctionName()
	var res RaftMemberListValue
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
