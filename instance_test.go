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
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestURLJoin(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("https://aslant.site/users/me", urlJoin("https://aslant.site", "/users/me"))
	assert.Equal("https://aslant.site/users/me", urlJoin("https://aslant.site/", "/users/me"))
	assert.Equal("https://aslant.site/users/me", urlJoin("https://aslant.site/", "https://aslant.site/users/me"))
	assert.Equal("http://aslant.site/users/me", urlJoin("https://aslant.site/", "http://aslant.site/users/me"))
}

func TestMergeConfig(t *testing.T) {
	assert := assert.New(t)
	headers := make(http.Header)
	headers.Set("a", "1")
	insConfig := &InstanceConfig{
		BaseURL: "https://aslant.site/",
		TransformRequest: []TransformRequest{
			func(body interface{}, headers http.Header) (data interface{}, err error) {
				return
			},
		},
		TransformResponse: []TransformResponse{
			func(body []byte, headers http.Header) (data []byte, err error) {
				return
			},
		},
		Headers: headers,
		Timeout: time.Second,
		Client:  new(http.Client),
		RequestInterceptors: []RequestInterceptor{
			func(config *Config) (err error) {
				return
			},
		},
		ResponseInterceptors: []ResponseInterceptor{
			func(resp *Response) (err error) {
				return
			},
		},
	}
	config := &Config{}
	mergeConfig(config, insConfig)
	assert.Equal(insConfig.BaseURL, config.BaseURL)
	assert.Equal(insConfig.TransformRequest, config.TransformRequest)
	assert.Equal(insConfig.TransformResponse, config.TransformResponse)
	assert.Equal(insConfig.Headers, config.Headers)
	assert.Equal(insConfig.Timeout, config.Timeout)
	assert.Equal(insConfig.Client, config.Client)
	assert.Equal(insConfig.RequestInterceptors, config.RequestInterceptors)
	assert.Equal(insConfig.ResponseInterceptors, config.ResponseInterceptors)
}

func TestIsNeedToTransformRequestBody(t *testing.T) {
	assert := assert.New(t)

	assert.True(isNeedToTransformRequestBody("POST"))
	assert.True(isNeedToTransformRequestBody("PATCH"))
	assert.True(isNeedToTransformRequestBody("PUT"))
	assert.False(isNeedToTransformRequestBody("GET"))
}

func TestNewRequest(t *testing.T) {

	assert := assert.New(t)

	postBody := []byte("abcd")
	tests := []struct {
		config   *Config
		method   string
		url      string
		route    string
		err      error
		postBody []byte
	}{
		{
			config: &Config{
				URL: "https://aslant.site/",
			},
			method: "GET",
			url:    "https://aslant.site/",
			route:  "/",
		},
		{
			config: &Config{
				URL: "https://aslant.site/user/:type",
				Params: map[string]string{
					"type": "me",
				},
			},
			method: "GET",
			url:    "https://aslant.site/user/me",
			route:  "/user/:type",
		},
		{
			config: &Config{
				URL: "https://aslant.site/",
				Query: url.Values{
					"type": []string{
						"vip",
					},
					"category": []string{
						"a",
						"b",
					},
				},
			},
			method: "GET",
			url:    "https://aslant.site/?category=a&category=b&type=vip",
			route:  "/",
		},
		{
			config: &Config{
				Method: "POST",
				URL:    "https://aslant.site/",
				Body:   postBody,
				TransformRequest: []TransformRequest{
					func(body interface{}, _ http.Header) (data interface{}, err error) {
						assert.Equal(postBody, body)
						return body, nil
					},
				},
			},
			method:   "POST",
			url:      "https://aslant.site/",
			postBody: postBody,
			route:    "/",
		},
		{
			config: &Config{
				Method: "测试",
			},
			err: errors.New(`net/http: invalid method "测试"`),
		},
	}

	for _, tt := range tests {
		req, err := newRequest(tt.config)
		assert.Equal(tt.err, err)
		if err == nil {
			assert.Equal(tt.method, req.Method)
			assert.Equal(tt.url, req.URL.String())
			assert.Equal(tt.route, tt.config.Route)
			if tt.postBody != nil {
				buf, _ := ioutil.ReadAll(req.Body)
				assert.Equal(tt.postBody, buf)
			}
		}
	}
}

func TestRequestInterceptor(t *testing.T) {

	assert := assert.New(t)
	customErr := errors.New("custom error")
	newCustomErr := errors.New("new custom error")
	deadlineExceededErr := errors.New(`Get "https://www.baidu.com/": context deadline exceeded`)
	// 1.13的出错没有双引号
	if runtime.Version() == "go1.13" {
		deadlineExceededErr = errors.New(`Get https://www.baidu.com/: context deadline exceeded`)
	}

	requestIDKey := "X-Request-ID"
	mockResp := &Response{
		Data: []byte("abc"),
	}
	mockGzipData := []byte("gzip data")
	mockAdapter := func(config *Config) (resp *Response, err error) {
		return mockResp, nil
	}
	tests := []struct {
		instanceConfig *InstanceConfig
		config         *Config
		resp           *Response
		err            error
		url            string
		requestID      string
	}{
		{
			config: &Config{
				URL: "https://aslant.site/users/:type",
				Query: url.Values{
					"category": []string{
						"a",
						"b",
					},
				},
				Params: map[string]string{
					"type": "me",
				},
				Headers: http.Header{
					requestIDKey: []string{
						"1",
					},
				},
				Timeout:     5 * time.Millisecond,
				enableTrace: true,
				Adapter:     mockAdapter,
			},
			resp:      mockResp,
			url:       "https://aslant.site/users/me?category=a&category=b",
			requestID: "1",
		},
		// request interceptors(error)
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				RequestInterceptors: []RequestInterceptor{
					func(config *Config) (err error) {
						if config.Route == "/error" {
							err = customErr
							return
						}
						config.Request.Header.Set(requestIDKey, "1")
						return
					},
				},
			},
			config: &Config{
				URL: "/error",
			},
			err: customErr,
		},
		// request interceptors (pass)
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				RequestInterceptors: []RequestInterceptor{
					func(config *Config) (err error) {
						if config.Route == "/error" {
							err = customErr
							return
						}
						config.Request.Header.Set(requestIDKey, "1")
						return
					},
				},
			},
			config: &Config{
				URL:     "/",
				Adapter: mockAdapter,
			},
			requestID: "1",
			url:       "https://aslant.site/",
			resp:      mockResp,
		},
		// transform response
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				Adapter: func(config *Config) (resp *Response, err error) {
					resp = &Response{}
					if config.Route == "/" {
						resp = mockResp
						return
					}

					return
				},
				TransformResponse: []TransformResponse{
					func(body []byte, headers http.Header) (data []byte, err error) {
						if bytes.Equal(body, mockResp.Data) {
							data = mockGzipData
							return
						}
						err = customErr
						return
					},
				},
			},
			config: &Config{
				URL: "/",
			},
			resp: &Response{
				Data: mockGzipData,
			},
			url: "https://aslant.site/",
		},

		// transform response (error)
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				Adapter: func(config *Config) (resp *Response, err error) {
					resp = &Response{}
					if config.Route == "/" {
						resp = mockResp
						return
					}

					return
				},
				TransformResponse: []TransformResponse{
					func(body []byte, headers http.Header) (data []byte, err error) {
						if bytes.Equal(body, mockResp.Data) {
							data = mockGzipData
							return
						}
						err = customErr
						return
					},
				},
			},
			config: &Config{
				URL: "/error",
			},
			err: customErr,
		},
		// response interceptors
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				Adapter: func(config *Config) (resp *Response, err error) {
					resp = &Response{
						Config: config,
					}
					return
				},
				ResponseInterceptors: []ResponseInterceptor{
					func(resp *Response) (err error) {
						if resp.Config.Route == "/error" {
							err = customErr
							return
						}
						resp.Data = mockResp.Data
						return
					},
				},
			},
			config: &Config{
				URL: "/",
			},
			resp: mockResp,
			url:  "https://aslant.site/",
		},
		// response interceptors (error)
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				Adapter: func(config *Config) (resp *Response, err error) {
					resp = &Response{
						Config: config,
					}
					return
				},
				ResponseInterceptors: []ResponseInterceptor{
					func(resp *Response) (err error) {
						if resp.Config.Route == "/error" {
							err = customErr
							return
						}
						resp.Data = mockResp.Data
						return
					},
				},
			},
			config: &Config{
				URL: "/error",
			},
			err: customErr,
		},
		// before new request
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				Adapter: mockAdapter,
				OnBeforeNewRequest: func(conf *Config) error {
					conf.BaseURL = "http://www.baidu.com"
					return nil
				},
			},
			config: &Config{
				URL: "/",
			},
			resp: mockResp,
			url:  "http://www.baidu.com/",
		},
		// on error
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				Adapter: func(config *Config) (resp *Response, err error) {
					err = customErr
					return
				},
				OnError: func(err error, config *Config) error {
					if err == customErr {
						return newCustomErr
					}
					return nil
				},
			},
			config: &Config{
				URL: "/",
			},
			err: newCustomErr,
		},
		// on done request success
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				Adapter: mockAdapter,
				OnDone: func(config *Config, resp *Response, err error) {
					// 设置至request id中
					if err != nil {
						config.Request.Header.Set(requestIDKey, "request fail")
					} else {
						config.Request.Header.Set(requestIDKey, "request success")
					}
				},
			},
			config: &Config{
				URL: "/",
			},
			url:       "https://aslant.site/",
			resp:      mockResp,
			requestID: "request success",
		},
		// on done request fail
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://aslant.site/",
				Adapter: func(config *Config) (resp *Response, err error) {
					err = customErr
					return
				},
				OnDone: func(config *Config, resp *Response, err error) {
					// 设置至request id中
					if err != nil {
						config.Request.Header.Set(requestIDKey, "request fail")
					} else {
						config.Request.Header.Set(requestIDKey, "request success")
					}
				},
			},
			config: &Config{
				URL: "/",
			},
			err:       customErr,
			requestID: "request fail",
		},
		// timeout
		{
			instanceConfig: &InstanceConfig{
				BaseURL: "https://www.baidu.com/",
				Timeout: time.Nanosecond,
			},
			config: &Config{
				URL: "/",
			},
			err: deadlineExceededErr,
		},
	}

	for _, tt := range tests {
		ins := NewInstance(tt.instanceConfig)
		resp, err := ins.Request(tt.config)
		if tt.err != nil || err != nil {
			assert.Equal(tt.err.Error(), err.Error())
		}
		assert.Equal(tt.requestID, tt.config.Request.Header.Get(requestIDKey))
		if tt.err == nil {
			if tt.resp != nil || resp != nil {
				assert.Equal(tt.resp.Data, resp.Data)
			}
			assert.Equal(tt.url, resp.Request.URL.String())
		}
	}
}

func TestRequest(t *testing.T) {
	assert := assert.New(t)
	mockResp := &Response{
		Data: []byte(`{
			"message": "hello world"
		}`),
	}
	mockPostData := map[string]string{
		"key": "1",
	}
	mockQuery := url.Values{
		"type": []string{
			"1",
		},
	}
	mockAdapter := func(config *Config) (resp *Response, err error) {
		if config.Query.Get("type") != "1" {
			return nil, errors.New("type is invalid")
		}

		if config.Method == "POST" ||
			config.Method == "PUT" ||
			config.Method == "PATCH" {
			buf, _ := ioutil.ReadAll(config.Request.Body)
			if string(buf) != `{"key":"1"}` {
				return nil, errors.New("request data is invalid")
			}
		}
		return mockResp, nil
	}

	ins := NewInstance(&InstanceConfig{
		Adapter: mockAdapter,
	})

	type MockResp struct {
		Message string
	}
	mockRespResult := &MockResp{
		Message: "hello world",
	}

	tests := []struct {
		fn     func() (*Response, error)
		err    error
		url    string
		resp   *Response
		method string
	}{
		// GetX
		{
			fn: func() (*Response, error) {
				return ins.GetX(context.Background(), "/", mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "GET",
		},
		// Get
		{
			fn: func() (*Response, error) {
				return ins.Get("/", mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "GET",
		},
		// DeleteX
		{
			fn: func() (*Response, error) {
				return ins.DeleteX(context.Background(), "/", mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "DELETE",
		},
		// Delete
		{
			fn: func() (*Response, error) {
				return ins.Delete("/", mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "DELETE",
		},
		// HeadX
		{
			fn: func() (*Response, error) {
				return ins.HeadX(context.Background(), "/", mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "HEAD",
		},
		// Head
		{
			fn: func() (*Response, error) {
				return ins.Head("/", mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "HEAD",
		},
		// OptionsX
		{
			fn: func() (*Response, error) {
				return ins.OptionsX(context.Background(), "/", mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "OPTIONS",
		},
		// Options
		{
			fn: func() (*Response, error) {
				return ins.Options("/", mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "OPTIONS",
		},
		// PostX
		{
			fn: func() (*Response, error) {
				return ins.PostX(context.Background(), "/", mockPostData, mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "POST",
		},
		// Post
		{
			fn: func() (*Response, error) {
				return ins.Post("/", mockPostData, mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "POST",
		},
		// PutX
		{
			fn: func() (*Response, error) {
				return ins.PutX(context.Background(), "/", mockPostData, mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "PUT",
		},
		// Put
		{
			fn: func() (*Response, error) {
				return ins.Put("/", mockPostData, mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "PUT",
		},
		// PatchX
		{
			fn: func() (*Response, error) {
				return ins.PatchX(context.Background(), "/", mockPostData, mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "PATCH",
		},
		// Patch
		{
			fn: func() (*Response, error) {
				return ins.Patch("/", mockPostData, mockQuery)
			},
			url:    "/?type=1",
			resp:   mockResp,
			method: "PATCH",
		},
	}

	for _, tt := range tests {
		resp, err := tt.fn()
		assert.Equal(tt.err, err)
		assert.Equal(tt.url, resp.Request.URL.RequestURI())
		assert.Equal(tt.resp, resp)
		assert.Equal(tt.method, resp.Request.Method)
	}

	enhanceTests := []struct {
		fn   func() (*MockResp, error)
		err  error
		resp *MockResp
	}{
		// EnhanceGetX
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhanceGetX(context.Background(), resp, "/", mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
		// EnhanceGet
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhanceGet(resp, "/", mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
		// EnhanceDeleteX
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhanceDeleteX(context.Background(), resp, "/", mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
		// EnhanceDelete
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhanceDelete(resp, "/", mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
		// EnhancePostX
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhancePostX(context.Background(), resp, "/", mockPostData, mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
		// EnhancePost
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhancePost(resp, "/", mockPostData, mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
		// EnhancePutX
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhancePutX(context.Background(), resp, "/", mockPostData, mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
		// EnhancePut
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhancePut(resp, "/", mockPostData, mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
		// EnhancePatchX
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhancePatchX(context.Background(), resp, "/", mockPostData, mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
		// EnhancePatch
		{
			fn: func() (*MockResp, error) {
				resp := &MockResp{}
				err := ins.EnhancePatch(resp, "/", mockPostData, mockQuery)
				return resp, err
			},
			resp: mockRespResult,
		},
	}
	for _, tt := range enhanceTests {
		resp, err := tt.fn()
		assert.Equal(tt.err, err)
		assert.Equal(tt.resp, resp)
	}
}

func TestMock(t *testing.T) {
	assert := assert.New(t)
	ins := NewInstance(nil)
	mockResp := &Response{}
	done := ins.Mock(mockResp)
	resp, err := ins.Get("/")
	assert.Nil(err)
	assert.Equal(mockResp, resp)
	done()

	done = ins.MultiMock(map[string]*Response{
		"/": mockResp,
	})
	defer done()
	resp, err = ins.Get("/")
	assert.Nil(err)
	assert.Equal(mockResp, resp)
}
