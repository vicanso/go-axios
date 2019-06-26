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
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
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
			func(body interface{}, headers http.Header) (data []byte, err error) {
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

func TestRequest(t *testing.T) {
	t.Run("mock adapter", func(t *testing.T) {
		assert := assert.New(t)
		mockAdapter := func(config *Config) (resp *Response, err error) {
			assert.Equal("/users/:type", config.Route)
			assert.Equal("https://aslant.site/users/me?a=1&category=a&category=b", config.URL)
			assert.Equal("GET", config.Method)
			assert.Equal(DefaultTransformResponse, config.TransformResponse)
			return
		}

		query := make(url.Values)
		query.Add("category", "a")
		query.Add("category", "b")
		ins := NewInstance(&InstanceConfig{
			BaseURL: "https://aslant.site/",
			Adapter: mockAdapter,
		})
		_, err := ins.Request(&Config{
			URL: "/users/:type?a=1",
			Params: map[string]string{
				"type": "me",
			},
			Query: query,
		})
		assert.Nil(err)
	})

	t.Run("on error", func(t *testing.T) {
		assert := assert.New(t)
		customErr := errors.New("abcd")
		mockAdapter := func(config *Config) (resp *Response, err error) {
			err = customErr
			return
		}

		ins := NewInstance(&InstanceConfig{
			BaseURL: "https://aslant.site/",
			Adapter: mockAdapter,
		})

		done := false
		_, err := ins.Request(&Config{
			URL: "/users/:type?a=1",
			Params: map[string]string{
				"type": "me",
			},
			OnError: func(err error, req *http.Request, resp *Response) {
				assert.Equal(customErr, err)
				done = true
			},
		})
		assert.NotNil(err)
		assert.True(done)
	})

	t.Run("request", func(t *testing.T) {
		assert := assert.New(t)
		defer gock.Off()
		gock.New("http://aslant.site").
			Get("/").
			MatchParam("a", "1").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		ins := NewInstance(nil)
		query := make(url.Values)
		query.Set("a", "1")
		resp, err := ins.Request(&Config{
			BaseURL: "http://aslant.site/",
			URL:     "/",
			Query:   query,
		})
		assert.Nil(err)
		assert.Equal(200, resp.Status)
	})

	t.Run("get", func(t *testing.T) {
		assert := assert.New(t)
		url := "https://aslant.site/"
		done := false
		mockAdapter := func(config *Config) (resp *Response, err error) {
			assert.Equal(url, config.URL)
			assert.Equal(http.MethodGet, config.Method)
			done = true
			return
		}

		ins := NewInstance(&InstanceConfig{
			Adapter: mockAdapter,
		})
		_, err := ins.Get(url)
		assert.Nil(err)
		assert.True(done)
	})

	t.Run("delete", func(t *testing.T) {
		assert := assert.New(t)
		url := "https://aslant.site/"
		done := false
		mockAdapter := func(config *Config) (resp *Response, err error) {
			assert.Equal(url, config.URL)
			assert.Equal(http.MethodDelete, config.Method)
			done = true
			return
		}

		ins := NewInstance(&InstanceConfig{
			Adapter: mockAdapter,
		})
		_, err := ins.Delete(url)
		assert.Nil(err)
		assert.True(done)
	})

	t.Run("head", func(t *testing.T) {
		assert := assert.New(t)
		url := "https://aslant.site/"
		done := false
		mockAdapter := func(config *Config) (resp *Response, err error) {
			assert.Equal(url, config.URL)
			assert.Equal(http.MethodHead, config.Method)
			done = true
			return
		}

		ins := NewInstance(&InstanceConfig{
			Adapter: mockAdapter,
		})
		_, err := ins.Head(url)
		assert.Nil(err)
		assert.True(done)
	})

	t.Run("options", func(t *testing.T) {
		assert := assert.New(t)
		url := "https://aslant.site/"
		done := false
		mockAdapter := func(config *Config) (resp *Response, err error) {
			assert.Equal(url, config.URL)
			assert.Equal(http.MethodOptions, config.Method)
			done = true
			return
		}

		ins := NewInstance(&InstanceConfig{
			Adapter: mockAdapter,
		})
		_, err := ins.Options(url)
		assert.Nil(err)
		assert.True(done)
	})

	t.Run("post", func(t *testing.T) {
		assert := assert.New(t)
		url := "https://aslant.site/"
		done := false
		data := []byte("abcd")
		mockAdapter := func(config *Config) (resp *Response, err error) {
			assert.Equal(url, config.URL)
			assert.Equal(http.MethodPost, config.Method)
			assert.Equal(data, config.Body)
			done = true
			return
		}

		ins := NewInstance(&InstanceConfig{
			Adapter: mockAdapter,
		})
		_, err := ins.Post(url, data)
		assert.Nil(err)
		assert.True(done)
	})

	t.Run("put", func(t *testing.T) {
		assert := assert.New(t)
		url := "https://aslant.site/"
		done := false
		data := []byte("abcd")
		mockAdapter := func(config *Config) (resp *Response, err error) {
			assert.Equal(url, config.URL)
			assert.Equal(http.MethodPut, config.Method)
			assert.Equal(data, config.Body)
			done = true
			return
		}

		ins := NewInstance(&InstanceConfig{
			Adapter: mockAdapter,
		})
		_, err := ins.Put(url, data)
		assert.Nil(err)
		assert.True(done)
	})

	t.Run("patch", func(t *testing.T) {
		assert := assert.New(t)
		url := "https://aslant.site/"
		done := false
		data := []byte("abcd")
		mockAdapter := func(config *Config) (resp *Response, err error) {
			assert.Equal(url, config.URL)
			assert.Equal(http.MethodPatch, config.Method)
			assert.Equal(data, config.Body)
			done = true
			return
		}

		ins := NewInstance(&InstanceConfig{
			Adapter: mockAdapter,
		})
		_, err := ins.Patch(url, data)
		assert.Nil(err)
		assert.True(done)
	})
}
