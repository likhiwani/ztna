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

package webapis

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"ztna-core/ztna/logtrace"
)

type GenericHttpHandler struct {
	HttpHandler http.Handler
	BindingKey  string
	ContextRoot string
}

func (spa *GenericHttpHandler) Binding() string {
	logtrace.LogWithFunctionName()
	return spa.BindingKey
}

func (spa *GenericHttpHandler) Options() map[interface{}]interface{} {
	logtrace.LogWithFunctionName()
	return nil
}

func (spa *GenericHttpHandler) RootPath() string {
	logtrace.LogWithFunctionName()
	return "/" + spa.BindingKey
}

func (spa *GenericHttpHandler) IsHandler(r *http.Request) bool {
	logtrace.LogWithFunctionName()
	return strings.HasPrefix(r.URL.Path, spa.ContextRoot) || strings.HasPrefix(r.URL.Path, "/assets")
}

func (spa *GenericHttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	logtrace.LogWithFunctionName()
	spa.HttpHandler.ServeHTTP(writer, request)
}

// Thanks to https://github.com/roberthodgen/spa-server
// Serve from a public directory with specific index
type spaHandler struct {
	content     string // The directory from which to serve
	contextRoot string // The context root to remove
	indexFile   string // The fallback/default file to serve
}

// Falls back to a supplied index (indexFile) when either condition is true:
// (1) Request (file) path is not found
// (2) Request path is a directory
// Otherwise serves the requested file.
func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logtrace.LogWithFunctionName()
	r.URL.Path = strings.TrimPrefix(r.URL.Path, h.contextRoot)
	p := filepath.Join(h.content, filepath.Clean(r.URL.Path))

	if info, err := os.Stat(p); err != nil {
		http.ServeFile(w, r, filepath.Join(h.content, h.indexFile))
		return
	} else if info.IsDir() {
		http.ServeFile(w, r, filepath.Join(h.content, h.indexFile))
		return
	}

	http.ServeFile(w, r, p)
}

// SpaHandler returns a request handler (http.Handler) that serves a single
// page application from a given public directory (location).
func SpaHandler(location string, contextRoot string, indexFile string) http.Handler {
	logtrace.LogWithFunctionName()
	return &spaHandler{location, contextRoot, indexFile}
}
