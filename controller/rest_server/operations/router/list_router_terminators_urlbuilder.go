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

package router

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"ztna-core/ztna/logtrace"
	"errors"
	"net/url"
	golangswaggerpaths "path"
	"strings"

	"github.com/go-openapi/swag"
)

// ListRouterTerminatorsURL generates an URL for the list router terminators operation
type ListRouterTerminatorsURL struct {
	ID string

	Filter *string
	Limit  *int64
	Offset *int64

	_basePath string
	// avoid unkeyed usage
	_ struct{}
}

// WithBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *ListRouterTerminatorsURL) WithBasePath(bp string) *ListRouterTerminatorsURL {
    logtrace.LogWithFunctionName()
	o.SetBasePath(bp)
	return o
}

// SetBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *ListRouterTerminatorsURL) SetBasePath(bp string) {
    logtrace.LogWithFunctionName()
	o._basePath = bp
}

// Build a url path and query string
func (o *ListRouterTerminatorsURL) Build() (*url.URL, error) {
    logtrace.LogWithFunctionName()
	var _result url.URL

	var _path = "/routers/{id}/terminators"

	id := o.ID
	if id != "" {
		_path = strings.Replace(_path, "{id}", id, -1)
	} else {
		return nil, errors.New("id is required on ListRouterTerminatorsURL")
	}

	_basePath := o._basePath
	if _basePath == "" {
		_basePath = "/fabric/v1"
	}
	_result.Path = golangswaggerpaths.Join(_basePath, _path)

	qs := make(url.Values)

	var filterQ string
	if o.Filter != nil {
		filterQ = *o.Filter
	}
	if filterQ != "" {
		qs.Set("filter", filterQ)
	}

	var limitQ string
	if o.Limit != nil {
		limitQ = swag.FormatInt64(*o.Limit)
	}
	if limitQ != "" {
		qs.Set("limit", limitQ)
	}

	var offsetQ string
	if o.Offset != nil {
		offsetQ = swag.FormatInt64(*o.Offset)
	}
	if offsetQ != "" {
		qs.Set("offset", offsetQ)
	}

	_result.RawQuery = qs.Encode()

	return &_result, nil
}

// Must is a helper function to panic when the url builder returns an error
func (o *ListRouterTerminatorsURL) Must(u *url.URL, err error) *url.URL {
    logtrace.LogWithFunctionName()
	if err != nil {
		panic(err)
	}
	if u == nil {
		panic("url can't be nil")
	}
	return u
}

// String returns the string representation of the path with query string
func (o *ListRouterTerminatorsURL) String() string {
    logtrace.LogWithFunctionName()
	return o.Must(o.Build()).String()
}

// BuildFull builds a full url with scheme, host, path and query string
func (o *ListRouterTerminatorsURL) BuildFull(scheme, host string) (*url.URL, error) {
    logtrace.LogWithFunctionName()
	if scheme == "" {
		return nil, errors.New("scheme is required for a full url on ListRouterTerminatorsURL")
	}
	if host == "" {
		return nil, errors.New("host is required for a full url on ListRouterTerminatorsURL")
	}

	base, err := o.Build()
	if err != nil {
		return nil, err
	}

	base.Scheme = scheme
	base.Host = host
	return base, nil
}

// StringFull returns the string representation of a complete url
func (o *ListRouterTerminatorsURL) StringFull(scheme, host string) string {
    logtrace.LogWithFunctionName()
	return o.Must(o.BuildFull(scheme, host)).String()
}
