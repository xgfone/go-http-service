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

// Predefine some variables
const (
	CharsetUTF8 = "charset=UTF-8"

	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + CharsetUTF8
	MIMEApplicationXML             = "application/xml"
	MIMEApplicationXMLCharsetUTF8  = MIMEApplicationXML + "; " + CharsetUTF8
	MIMEApplicationForm            = "application/x-www-form-urlencoded"
	MIMEMultipartForm              = "multipart/form-data"
)

// MIME slice types
var (
	mimeApplicationJSONs            = []string{MIMEApplicationJSON}
	mimeApplicationJSONCharsetUTF8s = []string{MIMEApplicationJSONCharsetUTF8}
	mimeApplicationXMLs             = []string{MIMEApplicationXML}
	mimeApplicationXMLCharsetUTF8s  = []string{MIMEApplicationXMLCharsetUTF8}
	mimeApplicationForms            = []string{MIMEApplicationForm}
	mimeMultipartForms              = []string{MIMEMultipartForm}
	mimeTextPlains                  = []string{"text/plain"}
)
