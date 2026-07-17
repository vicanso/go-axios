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
	"time"

	HT "github.com/vicanso/http-trace"
)

type (
	// ClientConfig is the config of client, it will be merged to
	// the config of each request of the client.
	ClientConfig struct {
		// BaseURL http request base url
		BaseURL string
		// TransformRequest transform request body
		TransformRequest []TransformRequest
		// TransformResponse transform response body
		TransformResponse []TransformResponse
		// Headers custom headers for all requests
		Headers http.Header
		// Timeout request timeout
		Timeout time.Duration
		// HTTPClient http client to send requests
		HTTPClient *http.Client
		// Adapter custom adapter
		Adapter Adapter
		// MaxConcurrency max concurrency for the client,
		// if it's lt 0, all requests will fail
		MaxConcurrency int32
		// EnableTrace enable http trace
		EnableTrace bool
		// RequestInterceptors request interceptor list
		RequestInterceptors []RequestInterceptor
		// ResponseInterceptors response interceptor list
		ResponseInterceptors []ResponseInterceptor
		// OnError error listener list
		OnError []OnError
		// OnDone done listener list
		OnDone []OnDone
		// OnBeforeRequest before request listener list
		OnBeforeRequest []OnBeforeRequest
		// JSONMarshal custom json marshal function
		JSONMarshal JSONMarshal
		// JSONUnmarshal custom json unmarshal function
		JSONUnmarshal JSONUnmarshal
	}

	// ClientOption sets an option of client config
	ClientOption func(config *ClientConfig)

	// Client is a http client with a client config,
	// it's concurrency safe.
	Client struct {
		config         ClientConfig
		mockAdapter    atomic.Pointer[Adapter]
		concurrency    atomic.Int32
		maxConcurrency atomic.Int32
	}

	// Mocker is a custom mock function for client
	Mocker func(config *Config) (*Response, error)
)

var ErrTooManyRequests = errors.New("too many request of the client")
var ErrNilResponse = errors.New("response is nil")

// WithBaseURL sets the base url of client
func WithBaseURL(baseURL string) ClientOption {
	return func(config *ClientConfig) {
		config.BaseURL = baseURL
	}
}

// WithTimeout sets the default timeout for all requests of client
func WithTimeout(d time.Duration) ClientOption {
	return func(config *ClientConfig) {
		config.Timeout = d
	}
}

// WithHTTPClient sets the http client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(config *ClientConfig) {
		config.HTTPClient = client
	}
}

// WithAdapter sets the adapter of client
func WithAdapter(adapter Adapter) ClientOption {
	return func(config *ClientConfig) {
		config.Adapter = adapter
	}
}

// WithHeader adds a header for all requests of client
func WithHeader(key, value string) ClientOption {
	return func(config *ClientConfig) {
		if config.Headers == nil {
			config.Headers = make(http.Header)
		}
		config.Headers.Add(key, value)
	}
}

// WithHeaders adds headers for all requests of client
func WithHeaders(headers http.Header) ClientOption {
	return func(config *ClientConfig) {
		if config.Headers == nil {
			config.Headers = make(http.Header)
		}
		for key, values := range headers {
			for _, value := range values {
				config.Headers.Add(key, value)
			}
		}
	}
}

// WithMaxConcurrency sets the max concurrency of client
func WithMaxConcurrency(value int32) ClientOption {
	return func(config *ClientConfig) {
		config.MaxConcurrency = value
	}
}

// WithEnableTrace enables http trace for all requests of client
func WithEnableTrace() ClientOption {
	return func(config *ClientConfig) {
		config.EnableTrace = true
	}
}

// WithTransformRequest sets the transform request functions of client
func WithTransformRequest(transforms ...TransformRequest) ClientOption {
	return func(config *ClientConfig) {
		config.TransformRequest = transforms
	}
}

// WithTransformResponse sets the transform response functions of client
func WithTransformResponse(transforms ...TransformResponse) ClientOption {
	return func(config *ClientConfig) {
		config.TransformResponse = transforms
	}
}

// WithRequestInterceptor appends request interceptors to client
func WithRequestInterceptor(interceptors ...RequestInterceptor) ClientOption {
	return func(config *ClientConfig) {
		config.RequestInterceptors = append(config.RequestInterceptors, interceptors...)
	}
}

// WithResponseInterceptor appends response interceptors to client
func WithResponseInterceptor(interceptors ...ResponseInterceptor) ClientOption {
	return func(config *ClientConfig) {
		config.ResponseInterceptors = append(config.ResponseInterceptors, interceptors...)
	}
}

// WithOnError appends error listeners to client
func WithOnError(listeners ...OnError) ClientOption {
	return func(config *ClientConfig) {
		config.OnError = append(config.OnError, listeners...)
	}
}

// WithOnDone appends done listeners to client
func WithOnDone(listeners ...OnDone) ClientOption {
	return func(config *ClientConfig) {
		config.OnDone = append(config.OnDone, listeners...)
	}
}

// WithOnBeforeRequest appends before request listeners to client
func WithOnBeforeRequest(listeners ...OnBeforeRequest) ClientOption {
	return func(config *ClientConfig) {
		config.OnBeforeRequest = append(config.OnBeforeRequest, listeners...)
	}
}

// WithJSONMarshal sets the json marshal function of client
func WithJSONMarshal(fn JSONMarshal) ClientOption {
	return func(config *ClientConfig) {
		config.JSONMarshal = fn
	}
}

// WithJSONUnmarshal sets the json unmarshal function of client
func WithJSONUnmarshal(fn JSONUnmarshal) ClientOption {
	return func(config *ClientConfig) {
		config.JSONUnmarshal = fn
	}
}

// NewClient creates a new client with options
func NewClient(opts ...ClientOption) *Client {
	config := ClientConfig{}
	for _, opt := range opts {
		opt(&config)
	}
	return NewClientWithConfig(config)
}

// NewClientWithConfig creates a new client with client config
func NewClientWithConfig(config ClientConfig) *Client {
	c := &Client{
		config: config,
	}
	c.maxConcurrency.Store(config.MaxConcurrency)
	return c
}

// GetConfig returns a copy of the client config
func (c *Client) GetConfig() ClientConfig {
	return c.config
}

// mergeConfig merges client config to request config
func (c *Client) mergeConfig(config *Config) {
	clientConfig := &c.config
	if clientConfig.EnableTrace {
		config.enableTrace = true
	}
	if config.BaseURL == "" {
		config.BaseURL = clientConfig.BaseURL
	}
	if config.TransformRequest == nil {
		config.TransformRequest = clientConfig.TransformRequest
	}
	if config.TransformResponse == nil {
		config.TransformResponse = clientConfig.TransformResponse
	}
	if config.Headers == nil {
		config.Headers = make(http.Header)
	}
	for key, values := range clientConfig.Headers {
		for _, value := range values {
			config.Headers.Add(key, value)
		}
	}

	if config.Timeout == 0 {
		config.Timeout = clientConfig.Timeout
	}
	if config.HTTPClient == nil {
		config.HTTPClient = clientConfig.HTTPClient
	}
	if config.Adapter == nil {
		config.Adapter = clientConfig.Adapter
	}
	if mock := c.mockAdapter.Load(); mock != nil {
		config.Adapter = *mock
	}
	if config.TransformResponse == nil {
		config.TransformResponse = DefaultTransformResponse
	}
	if config.TransformRequest == nil {
		config.TransformRequest = DefaultTransformRequest
	}
	if config.jsonMarshal == nil {
		config.jsonMarshal = clientConfig.JSONMarshal
	}
	if config.jsonUnmarshal == nil {
		config.jsonUnmarshal = clientConfig.JSONUnmarshal
	}

	// client的拦截器与监听函数在请求配置的之前执行
	config.RequestInterceptors = append(clientConfig.RequestInterceptors, config.RequestInterceptors...)
	config.ResponseInterceptors = append(clientConfig.ResponseInterceptors, config.ResponseInterceptors...)
	config.OnError = append(clientConfig.OnError, config.OnError...)
	config.OnDone = append(clientConfig.OnDone, config.OnDone...)
	config.OnBeforeRequest = append(clientConfig.OnBeforeRequest, config.OnBeforeRequest...)
}

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

	requestURL := config.GetURL()

	r, err := config.getRequestBody()
	if err != nil {
		return
	}

	return http.NewRequestWithContext(config.Context(), config.Method, requestURL, r)
}

func (c *Client) request(ctx context.Context, config *Config) (resp *Response, err error) {
	// 合并config必须放在第一步，因为有些事件是在client中配置
	c.mergeConfig(config)

	maxConcurrency := c.maxConcurrency.Load()
	if maxConcurrency < 0 {
		return nil, ErrRequestIsForbidden
	}
	config.Concurrency = c.concurrency.Add(1)
	defer c.concurrency.Add(-1)
	// 如果配置了最大请求数，而且当前请求大于最大请求数
	if maxConcurrency != 0 && config.Concurrency > maxConcurrency {
		return nil, ErrTooManyRequests
	}

	adapter := config.Adapter
	if adapter == nil {
		adapter = defaultAdapter
	}

	for _, fn := range config.OnBeforeRequest {
		err = fn(config)
		if err != nil {
			return
		}
	}
	// 如果在before request中已生成响应数据，则直接返回
	if config.Response != nil {
		config.Response.Config = config
		return config.Response, nil
	}

	if config.enableTrace {
		trace, ht := HT.NewClientTrace()
		ctx = httptrace.WithClientTrace(ctx, trace)
		config.HTTPTrace = ht
	}
	// 设置超时
	if config.Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}
	config.ctx = ctx

	req, err := newRequest(config)
	if err != nil {
		return
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
	if req.Header.Get(headerAccept) == "" {
		req.Header.Set(headerAccept, defaultAccept)
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
	if config.HTTPTrace != nil {
		config.HTTPTrace.Finish()
	}
	config.Response = resp
	if err != nil {
		return
	}
	if resp == nil {
		return nil, ErrNilResponse
	}
	resp.Config = config
	resp.Request = config.Request
	data := resp.Data
	// 响应数据的相关转换
	for _, fn := range config.TransformResponse {
		data, err = fn(config, data, resp.Headers)
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

// Do sends a http request with the config,
// the config will be modified by the client.
func (c *Client) Do(ctx context.Context, config *Config) (resp *Response, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	resp, err = c.request(ctx, config)
	if err != nil {
		for _, fn := range config.OnError {
			newErr := fn(err, config)
			if newErr != nil {
				err = newErr
				break
			}
		}
	}
	for _, fn := range config.OnDone {
		fn(config, resp, err)
	}
	return
}

func (c *Client) doMethod(ctx context.Context, method, url string, opts []RequestOption) (*Response, error) {
	config := &Config{
		URL:    url,
		Method: method,
	}
	for _, opt := range opts {
		opt(config)
	}
	return c.Do(ctx, config)
}

// Get sends a http get request
func (c *Client) Get(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	return c.doMethod(ctx, http.MethodGet, url, opts)
}

// Delete sends a http delete request
func (c *Client) Delete(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	return c.doMethod(ctx, http.MethodDelete, url, opts)
}

// Head sends a http head request
func (c *Client) Head(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	return c.doMethod(ctx, http.MethodHead, url, opts)
}

// Options sends a http options request
func (c *Client) Options(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	return c.doMethod(ctx, http.MethodOptions, url, opts)
}

// Post sends a http post request
func (c *Client) Post(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return c.doMethod(ctx, http.MethodPost, url, append(opts, Body(body)))
}

// Put sends a http put request
func (c *Client) Put(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return c.doMethod(ctx, http.MethodPut, url, append(opts, Body(body)))
}

// Patch sends a http patch request
func (c *Client) Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return c.doMethod(ctx, http.MethodPatch, url, append(opts, Body(body)))
}

// Upload sends a http post request with multipart form
func (c *Client) Upload(ctx context.Context, url string, form *MultipartForm, opts ...RequestOption) (*Response, error) {
	data, err := form.Bytes()
	if err != nil {
		return nil, err
	}
	opts = append(opts,
		Header(headerContentType, form.FormDataContentType()),
		Body(data),
	)
	return c.doMethod(ctx, http.MethodPost, url, opts)
}

// GetConcurrency returns the concurrency of client
func (c *Client) GetConcurrency() int32 {
	return c.concurrency.Load()
}

// SetMaxConcurrency sets the max concurrency of client
func (c *Client) SetMaxConcurrency(value int32) {
	c.maxConcurrency.Store(value)
}

// Mock mocks the response of client, it returns a function
// to remove the mock. It's used for testing.
func (c *Client) Mock(resp *Response, interceptors ...RequestInterceptor) (done func()) {
	return c.CustomMock(func(config *Config) (*Response, error) {
		for _, fn := range interceptors {
			err := fn(config)
			if err != nil {
				return nil, err
			}
		}
		return resp, nil
	})
}

// CustomMock sets a custom mock function of client, it returns
// a function to remove the mock. It's used for testing.
func (c *Client) CustomMock(fn Mocker) (done func()) {
	adapter := Adapter(func(config *Config) (*Response, error) {
		return fn(config)
	})
	c.mockAdapter.Store(&adapter)
	return func() {
		c.mockAdapter.Store(nil)
	}
}

// MultiMock mocks the responses of client by route, it returns
// a function to remove the mock. It's used for testing.
func (c *Client) MultiMock(multi map[string]*Response, interceptors ...RequestInterceptor) (done func()) {
	return c.CustomMock(func(config *Config) (*Response, error) {
		for _, fn := range interceptors {
			err := fn(config)
			if err != nil {
				return nil, err
			}
		}
		return multi[config.Route], nil
	})
}
