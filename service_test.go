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
	"net/http/httptest"
	"testing"
)

func TestService(t *testing.T) {
	svc := NewService()
	svc.Register("svc", func(c *Context) (err error) {
		var req struct {
			Name string `query:"Name" json:"Name"`
		}
		if err = c.Bind(&req); err != nil {
			return c.Failure(err)
		}

		return c.Success(req.Name)
	})

	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://127.0.0.1?Action=svc&Name=test", nil)
	if err != nil {
		t.Fatal(err)
	}

	svc.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expect status code '%d', but got '%d'", 200, rec.Code)
	}

	var result Response
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Errorf("failed to decode response by json: %v", err)
	} else if data, ok := result.Data.(string); !ok || data != "test" {
		t.Errorf("unexpect response '%+v'", result)
	}
}
