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
		UnmarshalData any
		Data          []byte
		Status        int
		Headers       http.Header
		Config        *Config
		Request       *http.Request
		// OriginalResponse original http response
		OriginalResponse *http.Response
	}
)

// JSON unmarshals the data of response to the value,
// it uses the json unmarshal function of config if set
func (resp *Response) JSON(v any) (err error) {
	if resp.Config != nil {
		err = resp.Config.unmarshalJSON(resp.Data, v)
	} else {
		err = defaultJSONUnmarshal(resp.Data, v)
	}
	if err == nil {
		resp.UnmarshalData = v
	}
	return
}

// OK returns true if the status of response is 2xx
func (resp *Response) OK() bool {
	return resp.Status >= http.StatusOK && resp.Status < http.StatusMultipleChoices
}
