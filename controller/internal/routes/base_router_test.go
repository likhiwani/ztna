/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package routes

import (
	"encoding/json"
	"reflect"
	"testing"
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/logtrace"

	"github.com/stretchr/testify/require"
)

var test = `
{
	"name" : "Foo",
	"terminatorStrategy" : "default",
	"configs" : [ "ssh-config", "ssh-server-config" ]
}
`

func Test_getFields(t *testing.T) {
	logtrace.LogWithFunctionName()
	assert := require.New(t)
	test2 := map[string]interface{}{
		"roleAttributes":     []string{"foo", "bar"},
		"name":               "Foo",
		"terminatorStrategy": "default",
		"configs":            []string{"ssh-config", "ssh-server-config"},
		"tags": map[string]interface{}{
			"foo": "bar",
			"nested": map[string]interface{}{
				"go": true,
			},
		}}

	test2Bytes, err := json.Marshal(test2)
	assert.NoError(err)

	tests := []struct {
		name    string
		body    []byte
		want    fields.UpdatedFieldsMap
		wantErr bool
	}{
		{
			name: "test",
			body: []byte(test),
			want: fields.UpdatedFieldsMap{
				"name":               struct{}{},
				"terminatorStrategy": struct{}{},
				"configs":            struct{}{},
			},
			wantErr: false,
		},
		{
			name: "test2",
			body: test2Bytes,
			want: fields.UpdatedFieldsMap{
				"name":               struct{}{},
				"terminatorStrategy": struct{}{},
				"roleAttributes":     struct{}{},
				"tags":               struct{}{},
				"configs":            struct{}{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := api.GetFields(tt.body)
			got.FilterMaps("tags")
			if (err != nil) != tt.wantErr {
				t.Errorf("getFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFields() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonFields_ConcatNestedNames(t *testing.T) {
	logtrace.LogWithFunctionName()
	tests := []struct {
		name string
		j    fields.UpdatedFieldsMap
		want fields.UpdatedFieldsMap
	}{
		{"test",
			fields.UpdatedFieldsMap{
				"Name":                  struct{}{},
				"This.Is.A.Longer.Name": struct{}{},
				"EgressRouter":          struct{}{},
			},
			fields.UpdatedFieldsMap{
				"Name":              struct{}{},
				"ThisIsALongerName": struct{}{},
				"EgressRouter":      struct{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.j.ConcatNestedNames()
			if !reflect.DeepEqual(tt.j, tt.want) {
				t.Errorf("ConcatNestedNames() got = %v, want %v", tt.j, tt.want)
			}
		})
	}
}
