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
	"encoding/json"
	"net/http"
)

var jsonBinder = JSONBinder()

// Binder is the interface to bind the value to v from ctx.
type Binder interface {
	// Bind parses the data from http.Request to v.
	//
	// Notice: v must be a non-nil pointer.
	Bind(v interface{}, r *http.Request) error
}

// BinderFunc is used to convert function to Binder.
type BinderFunc func(v interface{}, r *http.Request) error

// Bind implements the interface Binder.
func (f BinderFunc) Bind(v interface{}, r *http.Request) error { return f(v, r) }

// JSONBinder returns a Binder to decode and bind the request body with json.
//
// If ContentLength is equal to 0, it will do nothing.
func JSONBinder() Binder {
	return BinderFunc(func(v interface{}, r *http.Request) (err error) {
		if r.ContentLength > 0 {
			err = json.NewDecoder(r.Body).Decode(v)
		}
		return
	})
}
