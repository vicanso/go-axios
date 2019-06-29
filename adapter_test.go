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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestIsNeedToTransformRequestBody(t *testing.T) {
	assert := assert.New(t)
	assert.True(isNeedToTransformRequestBody("POST"))
	assert.True(isNeedToTransformRequestBody("PUT"))
	assert.True(isNeedToTransformRequestBody("PATCH"))
	assert.False(isNeedToTransformRequestBody("GET"))
}

func TestAdapter(t *testing.T) {
	t.Run("request get", func(t *testing.T) {
		assert := assert.New(t)
		mockData1 := []byte("mock data1")
		mockData2 := []byte("mock data2")
		defer gock.Off()
		gock.New("http://aslant.site").
			Get("/").
			MatchHeader("X-Request-ID", "123").
			MatchHeader("X-Interceptor", "true").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		headers := make(http.Header)
		headers.Set("X-Request-ID", "123")
		requestInterceptors := make([]RequestInterceptor, 0)
		requestInterceptors = append(requestInterceptors, func(config *Config) error {
			config.Request.Header.Set("X-Interceptor", "true")
			return nil
		})

		transformResponse := make([]TransformResponse, 0)
		transformResponse = append(transformResponse, func(body []byte, headers http.Header) ([]byte, error) {
			assert.NotEqual(mockData1, body)
			return mockData1, nil
		})

		responseInterceptors := make([]ResponseInterceptor, 0)
		responseInterceptors = append(responseInterceptors, func(resp *Response) error {
			assert.Equal(resp.Data, mockData1)
			resp.Data = mockData2
			return nil
		})

		config := &Config{
			enableTrace:          true,
			URL:                  "http://aslant.site/",
			Method:               "GET",
			Timeout:              time.Second,
			Headers:              headers,
			RequestInterceptors:  requestInterceptors,
			ResponseInterceptors: responseInterceptors,
			TransformResponse:    transformResponse,
		}
		resp, err := defaultAdapter(config)
		assert.Nil(err)
		assert.Equal(mockData2, resp.Data)
	})

	t.Run("new request error", func(t *testing.T) {
		assert := assert.New(t)
		_, err := defaultAdapter(&Config{
			Method: "测试",
		})
		assert.NotNil(err)
		assert.Equal(`message=net/http: invalid method "测试"`, err.Error())
	})

	t.Run("request interceptor error", func(t *testing.T) {
		assert := assert.New(t)
		customErr := errors.New("abcd")
		config := &Config{
			URL:    "http://aslant.site/",
			Method: "GET",
			RequestInterceptors: []RequestInterceptor{
				func(_ *Config) error {
					return customErr
				},
			},
		}
		_, e := defaultAdapter(config)
		assert.NotNil(e)
		err, ok := e.(*Error)
		assert.True(ok)
		assert.Equal(customErr, err.Err)
	})

	t.Run("request transform error", func(t *testing.T) {
		assert := assert.New(t)
		customErr := errors.New("abcd")
		config := &Config{
			Method: "POST",
			Body:   "abcd",
			TransformRequest: []TransformRequest{
				func(body interface{}, header http.Header) (interface{}, error) {
					return nil, customErr
				},
			},
		}

		_, e := defaultAdapter(config)
		assert.NotNil(e)
		err, ok := e.(*Error)
		assert.True(ok)
		assert.Equal(customErr, err.Err)
	})

	t.Run("request error", func(t *testing.T) {
		assert := assert.New(t)
		defer gock.Off()
		gock.New("http://aslant.site").
			Get("/users/me").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		config := &Config{
			URL:    "http://aslant.site/",
			Method: "GET",
		}
		_, e := defaultAdapter(config)
		assert.NotNil(e)
		err, ok := e.(*Error)
		assert.True(ok)
		assert.Nil(err.Config.Response)
	})

	t.Run("response transform error", func(t *testing.T) {
		assert := assert.New(t)
		customErr := errors.New("abcd")
		defer gock.Off()
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		config := &Config{
			URL:    "http://aslant.site/",
			Method: "GET",
			TransformResponse: []TransformResponse{
				func(body []byte, header http.Header) ([]byte, error) {
					return nil, customErr
				},
			},
		}
		_, e := defaultAdapter(config)
		assert.NotNil(e)
		err, ok := e.(*Error)
		assert.True(ok)
		assert.Equal(customErr, err.Err)
	})

	t.Run("response interceptor error", func(t *testing.T) {
		assert := assert.New(t)
		customErr := errors.New("abcd")
		defer gock.Off()
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		config := &Config{
			URL:    "http://aslant.site/",
			Method: "GET",
			ResponseInterceptors: []ResponseInterceptor{
				func(resp *Response) error {
					return customErr
				},
			},
		}
		_, e := defaultAdapter(config)
		assert.NotNil(e)
		err, ok := e.(*Error)
		assert.True(ok)
		assert.Equal(customErr, err.Err)
	})

	t.Run("post request", func(t *testing.T) {
		assert := assert.New(t)
		data := "hello world"
		defer gock.Off()
		gock.New("http://aslant.site").
			Post("/").
			BodyString("abcd").
			Reply(200).
			BodyString(data)
		config := &Config{
			URL:    "http://aslant.site/",
			Method: "POST",
			TransformRequest: []TransformRequest{
				convertRequestBody,
			},
			Body: []byte("abcd"),
		}
		resp, err := defaultAdapter(config)
		assert.Nil(err)
		assert.Equal(200, resp.Status)
		assert.Equal(data, string(resp.Data))
	})
}
