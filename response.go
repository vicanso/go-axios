// Copyright 2019 tree xie
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

package axios

import (
	"net/http"
)

type (
	// Response http response
	Response struct {
		Data    []byte
		Status  int
		Headers http.Header
		Config  *Config
		Request *http.Request
		// OriginalResponse original http response
		OriginalResponse *http.Response
	}
)

// JSON convert json data
func (resp *Response) JSON(v interface{}) (err error) {
	err = jsonUnmarshal(resp.Data, v)
	return
}
