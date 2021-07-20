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

package httpsvc_test

import (
	"fmt"
	"net/http"
	"sync/atomic"

	httpsvc "github.com/xgfone/go-http-service"
)

func Logger(name string) httpsvc.Middleware {
	return func(next httpsvc.Handler) httpsvc.Handler {
		var count uint64
		return func(c *httpsvc.Context) (err error) {
			index := atomic.AddUint64(&count, 1)
			fmt.Printf("[%d] log=%s, action=%s\n", index, name, c.Action)
			return next(c)
		}
	}
}

func ExampleService() {
	svc := httpsvc.NewService()

	// Add the global logging middleware.
	svc.Use(Logger("global"))

	// Rename the old service name "old_service1" to "service1".
	// So it is equal to call "service1" when calling "old_service1".
	svc.Mapping("old_service1", "service1")

	// Register the service "service1".
	svc.Register("service1", func(c *httpsvc.Context) error {
		return c.Success("service1")
	})

	// Register the service "service2" with the logger middleware
	// that only acts on the handler of "service2".
	svc.Register("service2", func(c *httpsvc.Context) (err error) {
		var req struct {
			Name string `query:"Name" json:"Name"`
		}

		if err = c.Bind(&req); err != nil {
			return c.Failure(err)
		}

		return c.Success(req.Name)
	}, Logger("service"))

	http.ListenAndServe("127.0.0.1:8080", svc)

	// ### Run Server:
	// $ go run main.go
	// [1] log=global, action=service1
	// [2] log=global, action=service2
	// [1] log=service, action=service2
	// [3] log=global, action=service2
	// [2] log=service, action=service2
	// [4] log=global, action=service2
	// [3] log=service, action=service2
	//
	//
	// ### Run Client:
	// $ curl -XGET 'http://127.0.0.1:8080/?Action=service1'
	// {"Data":"service1"}
	//
	// $ curl -XGET 'http://127.0.0.1:8080/?Action=old_service1'
	// {"Data":"service1"}
	//
	// $ curl -XGET 'http://127.0.0.1:8080/?Action=service2&Name=Aaron'
	// {"Data":"Aaron"}
	//
	// $ curl -XPOST 'http://127.0.0.1:8080/?Action=service2' -d '{"Name": "Aaron"}'
	// {"Data":"Aaron"}
	//
	// $ curl -XPOST 'http://127.0.0.1:8080/' -H 'X-Action: service2' -d '{"Name": "Aaron"}'
	// {"Data":"Aaron"}
}
