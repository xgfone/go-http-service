// Copyright 2021 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpsvc

import (
	"io"
	"net/http"
)

// ResponseWriter is the proxy of http.ResponseWriter.
type responseWriter struct {
	http.ResponseWriter

	Size   int64
	Wrote  bool
	Status int
}

// newResponse returns a new responseWriter.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, Status: http.StatusOK}
}

// WriteHeader implements http.ResponseWriter#WriteHeader().
func (r *responseWriter) WriteHeader(code int) {
	if !r.Wrote {
		r.Wrote = true
		r.Status = code
		r.ResponseWriter.WriteHeader(code)
	}
}

// Write implements http.ResponseWriter#Writer().
func (r *responseWriter) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}

	r.WriteHeader(http.StatusOK)
	n, err = r.ResponseWriter.Write(b)
	r.Size += int64(n)
	return
}

// WriteString implements io.StringWriter.
func (r *responseWriter) WriteString(s string) (n int, err error) {
	if len(s) == 0 {
		return
	}

	r.WriteHeader(http.StatusOK)
	n, err = io.WriteString(r.ResponseWriter, s)
	r.Size += int64(n)
	return
}

// Reset resets the response to the initialized and returns itself.
func (r *responseWriter) Reset(w http.ResponseWriter) {
	*r = responseWriter{ResponseWriter: w, Status: http.StatusOK}
}

// SetWriter resets the writer to w and return itself.
func (r *responseWriter) SetWriter(w http.ResponseWriter) { r.ResponseWriter = w }
