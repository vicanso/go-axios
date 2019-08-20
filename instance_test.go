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
	"errors"
	"net/http"
	"net/url"
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
	t.Run("new get request", func(t *testing.T) {
		assert := assert.New(t)
		method := "GET"
		url := "https://aslant.site/"
		req, err := newRequest(&Config{
			URL: url,
		})
		assert.Nil(err)
		assert.Equal(method, req.Method)
		assert.Equal(url, req.URL.String())
	})

	t.Run("request route", func(t *testing.T) {
		assert := assert.New(t)
		config := &Config{
			URL: "https://aslant.site/user/:type",
		}
		_, err := newRequest(config)
		assert.Nil(err)
		assert.Equal("/user/:type", config.Route)
	})

	t.Run("request params", func(t *testing.T) {
		assert := assert.New(t)
		config := &Config{
			URL: "https://aslant.site/user/:type",
			Params: map[string]string{
				"type": "me",
			},
		}
		_, err := newRequest(config)
		assert.Nil(err)
		assert.Equal("https://aslant.site/user/me", config.URL)
	})

	t.Run("request query", func(t *testing.T) {
		assert := assert.New(t)
		query := make(url.Values)
		query.Add("category", "a")
		query.Add("category", "b")
		config := &Config{
			URL:   "https://aslant.site/",
			Query: query,
		}
		_, err := newRequest(config)
		assert.Nil(err)
		assert.Equal("https://aslant.site/?category=a&category=b", config.URL)

		config = &Config{
			URL:   "https://aslant.site/?type=vip",
			Query: query,
		}
		_, err = newRequest(config)
		assert.Nil(err)
		assert.Equal("https://aslant.site/?type=vip&category=a&category=b", config.URL)
	})

	t.Run("invalid uri error", func(t *testing.T) {
		assert := assert.New(t)
		_, err := newRequest(&Config{
			Method: "测试",
		})
		assert.NotNil(`net/http: invalid method "测试"`, err.Error())
	})

	t.Run("new post request", func(t *testing.T) {
		assert := assert.New(t)
		method := "POST"
		postBody := []byte("abcd")
		url := "https://aslant.site/"
		req, err := newRequest(&Config{
			Method: method,
			URL:    url,
			Body:   postBody,
			TransformRequest: []TransformRequest{
				func(body interface{}, _ http.Header) (data interface{}, err error) {
					assert.Equal(postBody, body)
					return body, nil
				},
			},
		})
		assert.Nil(err)
		assert.Equal(method, req.Method)
		assert.Equal(url, req.URL.String())
		assert.NotNil(req.Body)
	})
}

func TestRequest(t *testing.T) {

	t.Run("requset", func(t *testing.T) {
		assert := assert.New(t)
		ins := NewInstance(nil)
		mockResp := &Response{}
		method := "GET"
		query := make(url.Values)
		query.Add("category", "a")
		query.Add("category", "b")
		header := make(http.Header)
		header.Set("X-Request-ID", "1")
		conf := &Config{
			URL:   "https://aslant.site/users/:type",
			Query: query,
			Params: map[string]string{
				"type": "me",
			},
			Headers:     header,
			Timeout:     5 * time.Millisecond,
			enableTrace: true,
		}

		conf.Adapter = func(config *Config) (resp *Response, err error) {
			req := config.Request
			assert.Equal(method, req.Method)
			assert.Equal("https://aslant.site/users/me?category=a&category=b", req.URL.String())
			assert.Equal("/users/:type", config.Route)
			assert.Equal("1", req.Header.Get("X-Request-ID"))
			assert.Equal(UserAgent, req.Header.Get(headerUserAgent))
			assert.Equal(defaultAcceptEncoding, req.Header.Get(headerAcceptEncoding))
			resp = mockResp
			return
		}
		resp, err := ins.Request(conf)
		assert.Nil(err)
		assert.Equal(mockResp, resp)
	})

	t.Run("request interceptors", func(t *testing.T) {
		assert := assert.New(t)
		customErr := errors.New("custom error")
		ins := NewInstance(&InstanceConfig{
			BaseURL: "https://aslant.site/",
			RequestInterceptors: []RequestInterceptor{
				func(config *Config) (err error) {
					if config.Route == "/error" {
						err = customErr
						return
					}
					config.Request.Header.Set("X-Request-ID", "1")
					return
				},
			},
			Adapter: func(config *Config) (resp *Response, err error) {
				resp = &Response{
					Request: config.Request,
				}
				return
			},
		})
		resp, err := ins.Get("/")
		assert.Nil(err)
		assert.Equal("1", resp.Request.Header.Get("X-Request-ID"))

		_, err = ins.Get("/error")
		assert.Equal(customErr, err.(*Error).Err)
	})

	t.Run("transform response", func(t *testing.T) {
		assert := assert.New(t)
		originalData := []byte("abcd")
		mockGzipData := []byte("gzip data")
		customErr := errors.New("custom error")
		ins := NewInstance(&InstanceConfig{
			BaseURL: "https://aslant.site/",
			Adapter: func(config *Config) (resp *Response, err error) {
				resp = &Response{}
				if config.Route == "/" {
					resp.Data = originalData
				}

				return
			},
			TransformResponse: []TransformResponse{
				func(body []byte, headers http.Header) (data []byte, err error) {
					if bytes.Equal(body, originalData) {
						data = mockGzipData
						return
					}
					err = customErr
					return
				},
			},
		})
		resp, err := ins.Get("/")
		assert.Nil(err)
		assert.Equal(mockGzipData, resp.Data)

		_, err = ins.Get("/error")
		assert.Equal(customErr, err.(*Error).Err)
	})

	t.Run("response interceptors", func(t *testing.T) {
		assert := assert.New(t)
		mockData := []byte("abcd")
		customErr := errors.New("custom error")
		ins := NewInstance(&InstanceConfig{
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
					resp.Data = mockData
					return
				},
			},
		})
		resp, err := ins.Get("/")
		assert.Nil(err)
		assert.Equal(mockData, resp.Data)

		_, err = ins.Get("/error")
		assert.Equal(customErr, err.(*Error).Err)
	})

	t.Run("on error", func(t *testing.T) {
		assert := assert.New(t)
		done := false
		customErr := errors.New("custom error")
		newCustomErr := errors.New("new custom error")
		ins := NewInstance(&InstanceConfig{
			BaseURL: "https://aslant.site/",
			Adapter: func(config *Config) (resp *Response, err error) {
				resp = &Response{
					Status: 200,
				}
				err = customErr
				return
			},
			OnError: func(err error, config *Config) error {
				done = true
				return newCustomErr
			},
		})

		_, err := ins.Get("/")
		assert.True(done)
		assert.Equal(newCustomErr, err)
	})

	t.Run("timeout", func(t *testing.T) {
		assert := assert.New(t)
		ins := NewInstance(&InstanceConfig{
			BaseURL: "https://aslant.site/",
			Timeout: time.Nanosecond,
		})
		_, err := ins.Get("/")
		e, ok := err.(*Error)
		assert.True(ok)
		assert.True(e.Timeout())
	})

	mockIns := NewInstance(&InstanceConfig{
		Adapter: func(config *Config) (resp *Response, err error) {
			resp = &Response{
				Config: config,
			}
			return
		},
	})
	t.Run("GET", func(t *testing.T) {
		assert := assert.New(t)
		resp, err := mockIns.Get("/")
		assert.Nil(err)
		assert.Equal("GET", resp.Config.Method)
		assert.Equal("/", resp.Config.URL)
	})
	t.Run("DELETE", func(t *testing.T) {
		assert := assert.New(t)
		resp, err := mockIns.Delete("/")
		assert.Nil(err)
		assert.Equal("DELETE", resp.Config.Method)
		assert.Equal("/", resp.Config.URL)
	})
	t.Run("HEAD", func(t *testing.T) {
		assert := assert.New(t)
		resp, err := mockIns.Head("/")
		assert.Nil(err)
		assert.Equal("HEAD", resp.Config.Method)
		assert.Equal("/", resp.Config.URL)
	})
	t.Run("OPTIONS", func(t *testing.T) {
		assert := assert.New(t)
		resp, err := mockIns.Options("/")
		assert.Nil(err)
		assert.Equal("OPTIONS", resp.Config.Method)
		assert.Equal("/", resp.Config.URL)
	})
	t.Run("POST", func(t *testing.T) {
		assert := assert.New(t)
		resp, err := mockIns.Post("/", map[string]string{
			"a": "1",
		})
		assert.Nil(err)
		assert.Equal("POST", resp.Config.Method)
		assert.Equal("/", resp.Config.URL)
	})
	t.Run("PUT", func(t *testing.T) {
		assert := assert.New(t)
		resp, err := mockIns.Put("/", map[string]string{
			"a": "1",
		})
		assert.Nil(err)
		assert.Equal("PUT", resp.Config.Method)
		assert.Equal("/", resp.Config.URL)
	})
	t.Run("PATCH", func(t *testing.T) {
		assert := assert.New(t)
		resp, err := mockIns.Patch("/", map[string]string{
			"a": "1",
		})
		assert.Nil(err)
		assert.Equal("PATCH", resp.Config.Method)
		assert.Equal("/", resp.Config.URL)
	})
}

func TestMock(t *testing.T) {
	assert := assert.New(t)
	ins := NewInstance(nil)
	mockResp := &Response{}
	done := ins.Mock(mockResp)
	defer done()
	resp, err := ins.Get("/")
	assert.Nil(err)
	assert.Equal(mockResp, resp)
}
