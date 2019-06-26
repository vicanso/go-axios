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
	"net/url"
	"strings"
)

type (
	// RequestInterceptor requset interceptor
	RequestInterceptor func(config *Config) (err error)
	// ResponseInterceptor response interceptor
	ResponseInterceptor func(resp *Response) (err error)

	// Instance instance of axios
	Instance struct {
		Config *InstanceConfig
	}
	// Response http response
	Response struct {
		Data    []byte
		Status  int
		Headers http.Header
		Config  *Config
		Request *http.Request
	}
)

// mergeConfig merge config
func mergeConfig(config *Config, insConfig *InstanceConfig) {
	if insConfig == nil {
		return
	}
	config.enableTrace = insConfig.EnableTrace
	if config.BaseURL == "" {
		config.BaseURL = insConfig.BaseURL
	}
	if config.TransformRequest == nil {
		config.TransformRequest = insConfig.TransformRequest
	}
	if config.TransformResponse == nil {
		config.TransformResponse = insConfig.TransformResponse
	}
	if config.Headers == nil {
		config.Headers = make(http.Header)
	}
	for key, values := range insConfig.Headers {
		for _, value := range values {
			config.Headers.Add(key, value)
		}
	}

	if config.Timeout == 0 {
		config.Timeout = insConfig.Timeout
	}
	if config.Client == nil {
		config.Client = insConfig.Client
	}
	if config.Adapter == nil {
		config.Adapter = insConfig.Adapter
	}
	if config.RequestInterceptors == nil {
		config.RequestInterceptors = insConfig.RequestInterceptors
	}
	if config.ResponseInterceptors == nil {
		config.ResponseInterceptors = insConfig.ResponseInterceptors
	}
	if config.OnError == nil {
		config.OnError = insConfig.OnError
	}
}

// NewInstance create a new instance
func NewInstance(config *InstanceConfig) *Instance {
	return &Instance{
		Config: config,
	}
}

func urlJoin(basicURL, url string) string {
	if basicURL == "" ||
		strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") {
		return url
	}
	if strings.HasSuffix(basicURL, "/") && strings.HasPrefix(url, "/") {
		return basicURL + url[1:]
	}
	return basicURL + url
}

// Request http request
func (ins *Instance) Request(config *Config) (resp *Response, err error) {
	if config.Headers == nil {
		config.Headers = make(http.Header)
	}
	mergeConfig(config, ins.Config)
	if config.Method == "" {
		config.Method = http.MethodGet
	}
	config.URL = urlJoin(config.BaseURL, config.URL)
	urlInfo, _ := url.Parse(config.URL)
	if urlInfo != nil {
		config.Route = urlInfo.Path
	}

	if config.Params != nil {
		for key, value := range config.Params {
			config.URL = strings.ReplaceAll(config.URL, ":"+key, value)
		}
	}

	if config.Query != nil {
		if strings.Contains(config.URL, "?") {
			config.URL += ("&" + config.Query.Encode())
		} else {
			config.URL += ("?" + config.Query.Encode())
		}
	}

	adapter := config.Adapter
	if adapter == nil {
		adapter = defaultAdapter
	}

	if config.TransformResponse == nil {
		config.TransformResponse = DefaultTransformResponse
	}

	if config.TransformRequest == nil {
		config.TransformRequest = DefaultTransformRequest
	}

	resp, err = adapter(config)
	if err != nil && config.OnError != nil {
		config.OnError(err, config.Request, config.Response)
	}
	return
}

// Get http get request
func (ins *Instance) Get(url string) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodGet,
	}
	return ins.Request(config)
}

// Delete http delete request
func (ins *Instance) Delete(url string) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodDelete,
	}
	return ins.Request(config)
}

// Head http head request
func (ins *Instance) Head(url string) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodHead,
	}
	return ins.Request(config)
}

// Options http options request
func (ins *Instance) Options(url string) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodOptions,
	}
	return ins.Request(config)
}

// Post http post request
func (ins *Instance) Post(url string, data interface{}) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodPost,
		Body:   data,
	}
	return ins.Request(config)
}

// Put http put request
func (ins *Instance) Put(url string, data interface{}) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodPut,
		Body:   data,
	}
	return ins.Request(config)
}

// Patch http patch request
func (ins *Instance) Patch(url string, data interface{}) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodPatch,
		Body:   data,
	}
	return ins.Request(config)
}
