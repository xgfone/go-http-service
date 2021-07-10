# Go HTTP Service [![Build Status](https://api.travis-ci.com/xgfone/go-http-service.svg?branch=master)](https://travis-ci.com/github/xgfone/go-http-service) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-http-service)](https://pkg.go.dev/github.com/xgfone/go-http-service) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-http-service/master/LICENSE)

Supply an action service framework based on http, supporting `Go1.5+`.

## Install
```shell
$ go get -u github.com/xgfone/go-http-service
```

## Usage
```go
package main

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

func main() {
	svc := httpsvc.NewService()

	// Add the global logging middleware.
	svc.Use(Logger("global"))

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
	// ### Run Client
	// $ curl -XGET 'http://127.0.0.1:8080/?Action=service1'
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
```
