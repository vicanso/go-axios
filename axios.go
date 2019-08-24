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

var (
	defaultIns *Instance
)

func init() {
	defaultIns = NewInstance(nil)
}

// Request http request by default instance
func Request(config *Config) (resp *Response, err error) {
	return defaultIns.Request(config)
}

// Get http get request by default instance
func Get(url string) (resp *Response, err error) {
	return defaultIns.Get(url)
}

// Delete http delete request by default instance
func Delete(url string) (resp *Response, err error) {
	return defaultIns.Delete(url)
}

// Head http head request by default instance
func Head(url string) (resp *Response, err error) {
	return defaultIns.Head(url)
}

// Options http options request by default instance
func Options(url string) (resp *Response, err error) {
	return defaultIns.Options(url)
}

// Post http post request by default instance
func Post(url string, data interface{}) (resp *Response, err error) {
	return defaultIns.Post(url, data)
}

// Put http put request by default instance
func Put(url string, data interface{}) (resp *Response, err error) {
	return defaultIns.Put(url, data)
}

// Patch http patch request by default instance
func Patch(url string, data interface{}) (resp *Response, err error) {
	return defaultIns.Patch(url, data)
}

// GetDefaultInstance get default instanc
func GetDefaultInstance() *Instance {
	return defaultIns
}
