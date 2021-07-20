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

// Package httpsvc supplies an action service framework based on http.
package httpsvc

import (
	"bytes"
	"net/http"
	"sync"
)

// Handler is the handler of the service.
type Handler func(*Context) error

// Middleware is the handler middleware.
type Middleware func(Handler) Handler

// Service is used to manager the services.
type Service struct {
	// NewContext is used to create the context.
	//
	// Default: NewContext
	NewContext func() *Context

	// GetAction is used to acquire the name of the service.
	//
	// Default: r.Header.Get("X-Action") or r.URL.Query().Get("Action")
	GetAction func(r *http.Request) (action string)

	// GetVersion is used to acquire the version of the requested service api.
	//
	// Default: r.Header.Get("X-Version")
	GetVersion func(r *http.Request) (version string)

	// GetRequestID is used to acquire the id of the request.
	//
	// Default: r.Header.Get("X-Request-Id")
	GetRequestID func(r *http.Request) (requestID string)

	mws     []Middleware
	handler Handler
	ctxpool sync.Pool
	bufpool sync.Pool

	lock     sync.RWMutex
	handlers map[string]Handler
	mappings map[string]string
}

// NewService returns a new Service.
func NewService() *Service {
	s := &Service{
		handlers: make(map[string]Handler),
		mappings: make(map[string]string),
	}

	s.handler = s.handleRequest
	s.bufpool.New = func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 2048))
	}
	s.ctxpool.New = func() interface{} {
		var ctx *Context
		if s.NewContext != nil {
			ctx = s.NewContext()
		} else {
			ctx = NewContext()
		}
		ctx.svc = s
		return ctx
	}

	return s
}

// Use registers the global middlewares that apply to all the services.
func (s *Service) Use(mws ...Middleware) {
	s.mws = append(s.mws, mws...)
	s.handler = s.handleRequest
	for _len := len(s.mws) - 1; _len >= 0; _len-- {
		s.handler = s.mws[_len](s.handler)
	}
}

// Register registers a service with the name and the handler.
func (s *Service) Register(name string, handler Handler, mws ...Middleware) {
	if name == "" {
		panic("Service.Register: the service name must not be empty")
	} else if handler == nil {
		panic("Service.Register: the service handler must not be empty")
	}

	for _len := len(mws) - 1; _len >= 0; _len-- {
		handler = mws[_len](handler)
	}

	s.lock.Lock()
	s.handlers[name] = handler
	s.lock.Unlock()
}

// Unregister unregisters the service by the name.
func (s *Service) Unregister(name string) {
	if name == "" {
		panic("Service.Unregister: the service name must not be empty")
	}

	s.lock.Lock()
	delete(s.handlers, name)
	s.lock.Unlock()
}

// Services returns the names of all the services.
func (s *Service) Services() (names []string) {
	s.lock.Lock()
	names = make([]string, 0, len(s.handlers))
	for name := range s.handlers {
		names = append(names, name)
	}
	s.lock.Unlock()
	return
}

// Mapping maps the name of the service from fromName to toName, that's,
// fromName is the alias of the name of the service named toName,
// and when calling the service named fromName, it will be forwarded
// to the service named toName to handle.
func (s *Service) Mapping(fromName, toName string) {
	if fromName == "" || toName == "" {
		panic("Service.Mapping: the service name must not be empty")
	}

	s.lock.Lock()
	s.mappings[fromName] = toName
	s.lock.Unlock()
}

// Mappings returns the mapping of the names of all the services.
func (s *Service) Mappings() map[string]string {
	s.lock.RLock()
	mappings := make(map[string]string, len(s.mappings))
	for k, v := range s.mappings {
		mappings[k] = v
	}
	s.lock.RUnlock()
	return mappings
}

func (s *Service) getHandler(name string) (handler Handler, ok bool) {
	s.lock.RLock()
	if handler, ok = s.handlers[name]; !ok {
		if name, ok = s.mappings[name]; ok {
			handler, ok = s.handlers[name]
		}
	}
	s.lock.RUnlock()
	return
}

// ServeHTTP implements the interface http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := s.ctxpool.Get().(*Context)
	c.SetReqResp(r, w)

	if s.GetAction != nil {
		c.Action = s.GetAction(c.req)
	} else if c.Action = c.GetReqHeader("X-Action"); c.Action == "" {
		c.Action = c.GetQuery("Action")
	}

	if s.GetVersion != nil {
		c.Version = s.GetVersion(c.req)
	} else {
		c.Version = c.GetReqHeader("X-Version")
	}

	if s.GetRequestID != nil {
		c.RequestID = s.GetRequestID(c.req)
	} else {
		c.RequestID = c.GetReqHeader("X-Request-Id")
	}

	if err := s.handler(c); !c.res.Wrote {
		c.Respond(nil, err)
	}

	c.reset()
	s.ctxpool.Put(c)
}

func (s *Service) handleRequest(c *Context) (err error) {
	if c.Action == "" {
		err = ErrInvalidAction.WithMessage("no action")
	} else if handler, ok := s.getHandler(c.Action); ok {
		err = handler(c)
	} else {
		err = ErrInvalidAction.WithMessage("invalid action '%s'", c.Action)
	}
	return
}
