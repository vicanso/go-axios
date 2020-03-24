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
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync/atomic"

	HT "github.com/vicanso/http-trace"
)

type (
	// RequestInterceptor requset interceptor
	RequestInterceptor func(config *Config) (err error)
	// ResponseInterceptor response interceptor
	ResponseInterceptor func(resp *Response) (err error)

	// Instance instance of axios
	Instance struct {
		Config      *InstanceConfig
		concurrency uint32
	}
)

var (
	needToTransformMethods = []string{
		http.MethodPost,
		http.MethodPatch,
		http.MethodPut,
	}
)

func isNeedToTransformRequestBody(method string) bool {
	for _, value := range needToTransformMethods {
		if value == method {
			return true
		}
	}
	return false
}

func newRequest(config *Config) (req *http.Request, err error) {
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

	var r io.Reader
	if config.Body != nil && isNeedToTransformRequestBody(config.Method) {
		data := config.Body
		for _, fn := range config.TransformRequest {
			buf, e := fn(data, config.Headers)
			if e != nil {
				err = e
				return
			}
			data = buf
		}
		r = bytes.NewReader(data.([]byte))
	}

	req, err = http.NewRequest(config.Method, config.URL, r)
	if err != nil {
		return
	}

	return
}

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
	if config.OnDone == nil {
		config.OnDone = insConfig.OnDone
	}
}

// NewInstance create a new instance
func NewInstance(config *InstanceConfig) *Instance {
	if config == nil {
		config = &InstanceConfig{}
	}
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

func (ins *Instance) request(config *Config) (resp *Response, err error) {
	config.Concurrency = atomic.AddUint32(&ins.concurrency, 1)
	defer atomic.AddUint32(&ins.concurrency, ^uint32(0))
	if config.Headers == nil {
		config.Headers = make(http.Header)
	}
	mergeConfig(config, ins.Config)

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

	req, err := newRequest(config)
	if err != nil {
		return
	}
	if config.enableTrace {
		ctx := config.Context
		if ctx == nil {
			ctx = context.Background()
		}
		trace, ht := HT.NewClientTrace()
		ctx = httptrace.WithClientTrace(ctx, trace)
		req = req.WithContext(ctx)
		defer ht.Finish()
		config.HTTPTrace = ht
		config.Context = ctx
	}
	// 设置超时
	if config.Timeout != 0 {
		ctx := config.Context
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, config.Timeout)
		defer cancel()
		config.Context = ctx
	}
	if config.Context != nil {
		req = req.WithContext(config.Context)
	}

	// 添加请求头
	for key, values := range config.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 设置默认的请求头
	if req.UserAgent() == "" {
		req.Header.Set(headerUserAgent, UserAgent)
	}
	if req.Header.Get(headerAcceptEncoding) == "" {
		req.Header.Set(headerAcceptEncoding, defaultAcceptEncoding)
	}

	config.Request = req

	// 请求前的相关拦截器调用
	for _, fn := range config.RequestInterceptors {
		err = fn(config)
		if err != nil {
			return
		}
	}

	resp, err = adapter(config)
	if err != nil {
		return
	}
	resp.Config = config
	resp.Request = config.Request
	config.Response = resp
	data := resp.Data
	// 响应数据的相关转换
	for _, fn := range config.TransformResponse {
		data, err = fn(data, resp.Headers)
		if err != nil {
			return
		}
	}
	resp.Data = data

	// 响应完成后的相关响应拦截器
	for _, fn := range config.ResponseInterceptors {
		err = fn(resp)
		if err != nil {
			return
		}
	}

	return
}

// Request http request
func (ins *Instance) Request(config *Config) (resp *Response, err error) {
	resp, err = ins.request(config)
	// TODO 添加done事件
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.Status
		}
		// 如果HTTP的响应码小于400，则出错是由于数据转换或拦截导致，
		// 错误码使用500
		if status < http.StatusBadRequest {
			status = http.StatusInternalServerError
		}
		err = CreateError(err, config, status)
	}
	if err != nil && config.OnError != nil {
		newErr := config.OnError(err, config)
		if newErr != nil {
			err = newErr
		}
	}
	if config.OnDone != nil {
		config.OnDone(config, resp, err)
	}
	return
}

// Get http get request
func (ins *Instance) Get(url string, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodGet,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.Request(config)
}

// Delete http delete request
func (ins *Instance) Delete(url string, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodDelete,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.Request(config)
}

// Head http head request
func (ins *Instance) Head(url string, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodHead,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.Request(config)
}

// Options http options request
func (ins *Instance) Options(url string, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodOptions,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.Request(config)
}

// Post http post request
func (ins *Instance) Post(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodPost,
		Body:   data,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.Request(config)
}

// Put http put request
func (ins *Instance) Put(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodPut,
		Body:   data,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.Request(config)
}

// Patch http patch request
func (ins *Instance) Patch(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		URL:    url,
		Method: http.MethodPatch,
		Body:   data,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.Request(config)
}

// Mock mock response
func (ins *Instance) Mock(resp *Response) (done func()) {
	originalAdapter := ins.Config.Adapter
	ins.Config.Adapter = func(_ *Config) (*Response, error) {
		return resp, nil
	}
	return func() {
		ins.Config.Adapter = originalAdapter
	}
}

// MultiMock multi mock response
func (ins *Instance) MultiMock(multi map[string]*Response) (done func()) {
	originalAdapter := ins.Config.Adapter
	ins.Config.Adapter = func(c *Config) (*Response, error) {
		resp := multi[c.Route]
		return resp, nil
	}
	return func() {
		ins.Config.Adapter = originalAdapter
	}
}
