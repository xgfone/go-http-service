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
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkServiceText(b *testing.B) {
	render := func(c *Context, r Response) error {
		return c.Text(200, "text/plain", "")
	}

	svc := NewService()
	svc.Register("service", func(c *Context) error { return c.Success(nil) })
	svc.NewContext = func() *Context {
		c := NewContext()
		c.Render = render
		return c
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://127.0.0.1", nil)
	req.Header.Set("X-Action", "service")
	if err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ServeHTTP(rec, req)
	}
}

func BenchmarkServiceJSON(b *testing.B) {
	svc := NewService()
	svc.Register("service", func(c *Context) error { return c.Success(nil) })

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://127.0.0.1", nil)
	req.Header.Set("X-Action", "service")
	if err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ServeHTTP(rec, req)
	}
}
