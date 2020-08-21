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
	"encoding/json"
	"net/url"
)

type (
	// JSONMarshal json marshal function type
	JSONMarshal func(interface{}) ([]byte, error)
	// JSONUnmarshal json unmarshal function type
	JSONUnmarshal func([]byte, interface{}) error
)

var (
	defaultIns *Instance
	// default json marshal
	jsonMarshal = json.Marshal
	// default json unmarshal
	jsonUnmarshal = json.Unmarshal
)

// SetJSONMarshal set json marshal function
func SetJSONMarshal(fn JSONMarshal) {
	jsonMarshal = fn
}

// SetJSONUnmarshal set json unmarshal function
func SetJSONUnmarshal(fn JSONUnmarshal) {
	jsonUnmarshal = fn
}

func init() {
	defaultIns = NewInstance(nil)
}

// Request http request by default instance
func Request(config *Config) (resp *Response, err error) {
	return defaultIns.Request(config)
}

// Get http get request by default instance
func Get(url string, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Get(url, query...)
}

// Delete http delete request by default instance
func Delete(url string, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Delete(url, query...)
}

// Head http head request by default instance
func Head(url string, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Head(url, query...)
}

// Options http options request by default instance
func Options(url string, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Options(url, query...)
}

// Post http post request by default instance
func Post(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Post(url, data, query...)
}

// Put http put request by default instance
func Put(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Put(url, data, query...)
}

// Patch http patch request by default instance
func Patch(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Patch(url, data, query...)
}

// GetDefaultInstance get default instanc
func GetDefaultInstance() *Instance {
	return defaultIns
}
