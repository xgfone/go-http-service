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
	"io"
	"net/http"
	"net/url"
)

func setContentType(header http.Header, ct string) {
	if ct != "" {
		switch ct {
		case MIMEApplicationJSON:
			header["Content-Type"] = mimeApplicationJSONs
		case MIMEApplicationJSONCharsetUTF8:
			header["Content-Type"] = mimeApplicationJSONCharsetUTF8s
		case MIMEApplicationXML:
			header["Content-Type"] = mimeApplicationXMLs
		case MIMEApplicationXMLCharsetUTF8:
			header["Content-Type"] = mimeApplicationXMLCharsetUTF8s
		case MIMEApplicationForm:
			header["Content-Type"] = mimeApplicationForms
		case MIMEMultipartForm:
			header["Content-Type"] = mimeMultipartForms
		case "text/plain":
			header["Content-Type"] = mimeTextPlains
		default:
			header.Set("Content-Type", ct)
		}
	}
}

// Response represents a response result.
type Response struct {
	RequestId string      `json:",omitempty" xml:",omitempty"`
	Error     Error       `json:",omitempty" xml:",omitempty"`
	Data      interface{} `json:",omitempty" xml:",omitempty"`
}

type Context struct {
	// Action is the name of the service.
	Action string

	// Version is the version of the service api.
	Version string

	// RequestID is the unique id indicating the request.
	RequestID string

	// Data is used to save the context data during handling the request.
	Data interface{}

	// BindBody is used to bind the request body to data.
	//
	// Default: JSONBinder()
	BindBody Binder

	// BindQuery is used to bind the request query to data.
	//
	// Default: use BindURLValues(data, query, "query")
	BindQuery func(data interface{}, query url.Values) error

	// SetDefault is used to set the data to the default if it is ZERO.
	//
	// If data is a struct, set the fields of the struct to the default if ZERO.
	//
	// Default: nil
	SetDefault func(data interface{}) error

	// Validate is used to validate whether data is valid.
	//
	// Default: nil
	Validate func(data interface{}) error

	// Render is used to render the response.
	//
	// Default: use c.JSON(r)
	Render func(c *Context, r Response) error

	req *http.Request
	res *responseWriter

	query url.Values
}

// NewContext returns a new Context.
func NewContext() *Context { return &Context{res: newResponseWriter(nil)} }

func (c *Context) reset() {
	c.Data = nil
	c.req, c.query = nil, nil
	c.res.Reset(nil)
}

// StatusCode returns the status code of the response.
func (c *Context) StatusCode() int { return c.res.Status }

// IsResponded reports whether the response is sent.
func (c *Context) IsResponded() bool { return c.res.Wrote }

// Request returns the inner Request.
func (c *Context) Request() *http.Request { return c.req }

// SetRequest resets the request to req.
func (c *Context) SetRequest(req *http.Request) { c.req = req }

// ResponseWriter returns the underlying http.ResponseWriter.
func (c *Context) ResponseWriter() http.ResponseWriter {
	return c.res.ResponseWriter
}

// SetResponseWriter resets the response to resp.
func (c *Context) SetResponseWriter(resp http.ResponseWriter) {
	c.res.ResponseWriter = resp
}

// SetReqResp is equal to the union of SetRequest and SetResponseWriter.
func (c *Context) SetReqResp(req *http.Request, resp http.ResponseWriter) {
	c.req, c.res.ResponseWriter = req, resp
}

// Blob sends the binary data to the client with status code and content type.
func (c *Context) Blob(code int, contentType string, data []byte) (err error) {
	setContentType(c.res.Header(), contentType)
	c.res.WriteHeader(code)
	if len(data) > 0 {
		_, err = c.res.Write(data)
	}
	return
}

// Text sends the string text to the client with status code and content type.
func (c *Context) Text(code int, contentType string, data string) (err error) {
	setContentType(c.res.Header(), contentType)
	c.res.WriteHeader(code)
	if len(data) > 0 {
		_, err = c.res.WriteString(data)
	}
	return
}

// Stream sends the data from the stream to the client with status code
// and content type.
func (c *Context) Stream(code int, contentType string, r io.Reader) (err error) {
	setContentType(c.res.Header(), contentType)
	c.res.WriteHeader(code)
	_, err = io.CopyBuffer(c.res, r, make([]byte, 2048))
	return
}

// JSON encodes the data with the json encoder, then responds to the client
// with the status code 200.
func (c *Context) JSON(data interface{}) error {
	setContentType(c.res.Header(), MIMEApplicationJSONCharsetUTF8)
	c.res.WriteHeader(200)
	return json.NewEncoder(c.res).Encode(data)
}

// Respond sends the response as Response.
//
// If Render isn't nil, use it to render the response. Or use c.JSON instead.
func (c *Context) Respond(data interface{}, err error) error {
	var _err Error
	switch e := err.(type) {
	case nil:
	case Error:
		_err = Error{e.Code, e.Message}
	case interface{ CodeError() Error }:
		_err = e.CodeError()
	default:
		_err = ErrServerError.WithMessage(err.Error())
	}

	if c.Render != nil {
		return c.Render(c, Response{RequestId: c.RequestID, Error: _err, Data: data})
	}

	type Resp struct {
		RequestId string      `json:",omitempty" xml:",omitempty"`
		Error     error       `json:",omitempty" xml:",omitempty"`
		Data      interface{} `json:",omitempty" xml:",omitempty"`
	}

	if _err.Code == "" {
		return c.JSON(Resp{RequestId: c.RequestID, Data: data})
	}
	return c.JSON(Resp{RequestId: c.RequestID, Error: _err, Data: data})
}

// Success is equal to c.Respond("", data, nil).
func (c *Context) Success(data interface{}) error { return c.Respond(data, nil) }

// Failure is equal to c.Respond("", nil, err).
func (c *Context) Failure(err error) error { return c.Respond(nil, err) }

// Query parses and returns the query of the request.
func (c *Context) Query() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query
}

// GetQuery is equal to c.Query().Get(key).
func (c *Context) GetQuery(key string) string { return c.Query().Get(key) }

// GetReqHeader is equal to c.Request().Header.Get(key).
func (c *Context) GetReqHeader(key string) string { return c.req.Header.Get(key) }

// SetRespHeader is equal to c.ResponseWriter().Header().Set(key, value).
func (c *Context) SetRespHeader(key, value string) { c.res.Header().Set(key, value) }

// Bind is used to bind the request to v, then set the default
// and validate the data.
//
// If the method is "GET", use c.BindQuery.
// If the method is "POST", use c.BindBody.
func (c *Context) Bind(v interface{}) (err error) {
	switch c.req.Method {
	case http.MethodGet:
		if c.BindQuery != nil {
			err = c.BindQuery(v, c.Query())
		} else {
			err = BindURLValues(v, c.Query(), "query")
		}
	case http.MethodPost:
		if c.BindBody != nil {
			err = c.BindBody.Bind(v, c.req)
		} else {
			err = jsonBinder.Bind(v, c.req)
		}
	default:
		return ErrUnsupportedProtocol.WithMessage("unsupported method '%s'", c.req.Method)
	}

	if err == nil {
		if c.SetDefault != nil {
			err = c.SetDefault(v)
		}

		if err == nil && c.Validate != nil {
			err = c.Validate(v)
		}
	}

	switch err.(type) {
	case nil:
	case Error:
	default:
		err = ErrInvalidParameter.WithMessage(err.Error())
	}

	return
}
