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
	"context"
	"errors"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync/atomic"

	HT "github.com/vicanso/http-trace"
)

type (
	// RequestInterceptor request interceptor
	RequestInterceptor func(config *Config) (err error)
	// ResponseInterceptor response interceptor
	ResponseInterceptor func(resp *Response) (err error)

	// Instance instance of axios
	Instance struct {
		Config      *InstanceConfig
		concurrency uint32
	}
)

var ErrTooManyRequests = errors.New("too many request of the instance")

func newRequest(config *Config) (req *http.Request, err error) {
	if config.Method == "" {
		config.Method = http.MethodGet
	}
	if config.Route == "" {
		urlInfo, _ := url.Parse(config.URL)
		if urlInfo != nil {
			config.Route = urlInfo.Path
		}
	}

	url := config.GetURL()

	r, err := config.getRequestBody()
	if err != nil {
		return
	}

	req, err = http.NewRequest(config.Method, url, r)
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
	if insConfig.EnableTrace {
		config.enableTrace = insConfig.EnableTrace
	}
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
	if config.OnBeforeNewRequest == nil {
		config.OnBeforeNewRequest = insConfig.OnBeforeNewRequest
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

func (ins *Instance) request(config *Config) (resp *Response, err error) {
	// 合并config必须放在第一步，因为有些事件是在instance中生成
	mergeConfig(config, ins.Config)

	if ins.Config.MaxConcurrency < 0 {
		return nil, ErrRequestIsForbidden
	}
	config.Concurrency = atomic.AddUint32(&ins.concurrency, 1)
	defer atomic.AddUint32(&ins.concurrency, ^uint32(0))
	// 如果配置了最大请求数，而且当前请求大于最大请求数
	if ins.Config.MaxConcurrency != 0 && int32(config.Concurrency) > ins.Config.MaxConcurrency {
		err = ErrTooManyRequests
		return
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

	if config.OnBeforeNewRequest != nil {
		err = config.OnBeforeNewRequest(config)
		if err != nil {
			return
		}
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
	// 如果未设置Accept，则设置默认值 application/json, text/plain, */*
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accpet", "application/json, text/plain, */*")
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
	config.Response = resp
	if err != nil {
		return
	}
	resp.Config = config
	resp.Request = config.Request
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

// doRequest do http request
func (ins *Instance) doRequest(config *Config, result interface{}) (resp *Response, err error) {
	resp, err = ins.request(config)
	if err != nil && config.OnError != nil {
		newErr := config.OnError(err, config)
		if newErr != nil {
			err = newErr
		}
	}
	// 如果没有出错，有响应结果，而且指定了result(需要unmarshal)
	if err == nil &&
		resp != nil &&
		result != nil {
		err = resp.JSON(result)
	}
	if config.OnDone != nil {
		config.OnDone(config, resp, err)
	}
	return
}

// GetConcurrency get concurrency of instance
func (ins *Instance) GetConcurrency() uint32 {
	return atomic.LoadUint32(&ins.concurrency)
}

// SetMaxConcurrency sets max concurrency for instance
func (ins *Instance) SetMaxConcurrency(value int32) {
	atomic.StoreInt32(&ins.Config.MaxConcurrency, value)
}

// Request http request
func (ins *Instance) Request(config *Config) (resp *Response, err error) {
	resp, err = ins.doRequest(config, nil)
	return
}

// EnhanceRequest http request and unmarshal response to struct
func (ins *Instance) EnhanceRequest(result interface{}, config *Config) (err error) {
	_, err = ins.doRequest(config, result)
	if err != nil {
		return
	}
	return
}

// getX http get request with context
func (ins *Instance) getX(context context.Context, result interface{}, url string, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		Context: context,
		URL:     url,
		Method:  http.MethodGet,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.doRequest(config, result)
}

// GetX http get request with context
func (ins *Instance) GetX(context context.Context, url string, query ...url.Values) (resp *Response, err error) {
	return ins.getX(context, nil, url, query...)
}

// EnhanceGetX http get request with context and unmarshal response to struct
func (ins *Instance) EnhanceGetX(context context.Context, result interface{}, url string, query ...url.Values) (err error) {
	_, err = ins.getX(context, result, url, query...)
	if err != nil {
		return
	}
	return
}

// Get http get request
func (ins *Instance) Get(url string, query ...url.Values) (resp *Response, err error) {
	return ins.getX(context.Background(), nil, url, query...)
}

// EnhanceGetX http get request with context and unmarshal response to struct
func (ins *Instance) EnhanceGet(result interface{}, url string, query ...url.Values) (err error) {
	_, err = ins.getX(context.Background(), result, url, query...)
	if err != nil {
		return
	}
	return
}

// deleteX http delete request with context
func (ins *Instance) deleteX(context context.Context, result interface{}, url string, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		Context: context,
		URL:     url,
		Method:  http.MethodDelete,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.doRequest(config, result)
}

// DeleteX http delete request with context
func (ins *Instance) DeleteX(context context.Context, url string, query ...url.Values) (resp *Response, err error) {
	return ins.deleteX(context, nil, url, query...)
}

// EnhanceDeleteX http delete request with context and unmarshal response to struct
func (ins *Instance) EnhanceDeleteX(context context.Context, result interface{}, url string, query ...url.Values) (err error) {
	_, err = ins.deleteX(context, result, url, query...)
	if err != nil {
		return
	}
	return
}

// Delete http delete request
func (ins *Instance) Delete(url string, query ...url.Values) (resp *Response, err error) {
	return ins.deleteX(context.Background(), nil, url, query...)
}

// EnhanceDelete http delete request and unmarshal response to struct
func (ins *Instance) EnhanceDelete(result interface{}, url string, query ...url.Values) (err error) {
	_, err = ins.deleteX(context.Background(), result, url, query...)
	if err != nil {
		return
	}
	return
}

// headX http head request with context
func (ins *Instance) headX(context context.Context, result interface{}, url string, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		URL:     url,
		Method:  http.MethodHead,
		Context: context,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.doRequest(config, result)
}

// HeadX http head request with context
func (ins *Instance) HeadX(context context.Context, url string, query ...url.Values) (resp *Response, err error) {
	return ins.headX(context, nil, url, query...)
}

// Head http head request
func (ins *Instance) Head(url string, query ...url.Values) (resp *Response, err error) {
	return ins.headX(context.Background(), nil, url, query...)
}

// optionsX http options request with context
func (ins *Instance) optionsX(context context.Context, result interface{}, url string, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		Context: context,
		URL:     url,
		Method:  http.MethodOptions,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.doRequest(config, result)
}

// OptionsX http options request with context
func (ins *Instance) OptionsX(context context.Context, url string, query ...url.Values) (resp *Response, err error) {
	return ins.optionsX(context, nil, url, query...)
}

// Options http options request
func (ins *Instance) Options(url string, query ...url.Values) (resp *Response, err error) {
	return ins.optionsX(context.Background(), nil, url, query...)
}

// postX http post request with context
func (ins *Instance) postX(context context.Context, result interface{}, url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		Context: context,
		URL:     url,
		Method:  http.MethodPost,
		Body:    data,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.doRequest(config, result)
}

// PostX http post request with context
func (ins *Instance) PostX(context context.Context, url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return ins.postX(context, nil, url, data, query...)
}

// EnhancePostX http post request with context and unmarshal response to struct
func (ins *Instance) EnhancePostX(context context.Context, result interface{}, url string, data interface{}, query ...url.Values) (err error) {
	_, err = ins.postX(context, result, url, data, query...)
	if err != nil {
		return
	}
	return
}

// Post http post request
func (ins *Instance) Post(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return ins.postX(context.Background(), nil, url, data, query...)
}

// EnhancePost http post request and unmarshal response to struct
func (ins *Instance) EnhancePost(result interface{}, url string, data interface{}, query ...url.Values) (err error) {
	_, err = ins.postX(context.Background(), result, url, data, query...)
	if err != nil {
		return
	}
	return
}

// putX http put request with context
func (ins *Instance) putX(context context.Context, result interface{}, url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		Context: context,
		URL:     url,
		Method:  http.MethodPut,
		Body:    data,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.doRequest(config, result)
}

// PutX http put request with context
func (ins *Instance) PutX(context context.Context, url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return ins.putX(context, nil, url, data, query...)
}

// EnhancePutX http put request with context and unmarshal response to struct
func (ins *Instance) EnhancePutX(context context.Context, result interface{}, url string, data interface{}, query ...url.Values) (err error) {
	_, err = ins.putX(context, result, url, data, query...)
	if err != nil {
		return
	}
	return
}

// Put http put request
func (ins *Instance) Put(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return ins.putX(context.Background(), nil, url, data, query...)
}

// EnhancePut http put request and unmarshal response to struct
func (ins *Instance) EnhancePut(result interface{}, url string, data interface{}, query ...url.Values) (err error) {
	_, err = ins.putX(context.Background(), result, url, data, query...)
	if err != nil {
		return
	}
	return
}

// patchX http patch request with context
func (ins *Instance) patchX(context context.Context, result interface{}, url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	config := &Config{
		Context: context,
		URL:     url,
		Method:  http.MethodPatch,
		Body:    data,
	}
	if len(query) != 0 {
		config.Query = query[0]
	}
	return ins.doRequest(config, result)
}

// PatchX http patch request with context
func (ins *Instance) PatchX(context context.Context, url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return ins.patchX(context, nil, url, data, query...)
}

// Patch http patch request
func (ins *Instance) Patch(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return ins.patchX(context.Background(), nil, url, data, query...)
}

// EnhancePatchX http patch request with context and unmarshal response to struct
func (ins *Instance) EnhancePatchX(context context.Context, result interface{}, url string, data interface{}, query ...url.Values) (err error) {
	_, err = ins.patchX(context, result, url, data, query...)
	if err != nil {
		return
	}
	return
}

// EnhancePatch http patch request and unmarshal response to struct
func (ins *Instance) EnhancePatch(result interface{}, url string, data interface{}, query ...url.Values) (err error) {
	_, err = ins.patchX(context.Background(), result, url, data, query...)
	if err != nil {
		return
	}
	return
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

// AppendRequestInterceptor appends request interceptor to instance
func (ins *Instance) AppendRequestInterceptor(fn RequestInterceptor) {
	ins.Config.RequestInterceptors = append(ins.Config.RequestInterceptors, fn)
}

// PrependRequestInterceptor prepends request interceptor to instance
func (ins *Instance) PrependRequestInterceptor(fn RequestInterceptor) {
	ins.Config.RequestInterceptors = append([]RequestInterceptor{
		fn,
	}, ins.Config.RequestInterceptors...)
}

// AppendResponseInterceptor appends response interceptor to instance
func (ins *Instance) AppendResponseInterceptor(fn ResponseInterceptor) {
	ins.Config.ResponseInterceptors = append(ins.Config.ResponseInterceptors, fn)
}

// PrependResponseInterceptor prepends response interceptor to instance
func (ins *Instance) PrependResponseInterceptor(fn ResponseInterceptor) {
	ins.Config.ResponseInterceptors = append([]ResponseInterceptor{
		fn,
	}, ins.Config.ResponseInterceptors...)
}
