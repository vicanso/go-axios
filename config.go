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
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cast"
	HT "github.com/vicanso/http-trace"
)

type (
	// OnError on error function
	OnError func(err error, config *Config) (newErr error)
	// Config http request config
	Config struct {
		Request  *http.Request
		Response *Response

		// Route the request route
		Route string
		// URL the request url
		URL string
		// Method http request method, default is `get`
		Method string
		// BaseURL http request base url
		BaseURL string
		// TransformRequest transform requset body
		TransformRequest []TransformRequest
		// TransformResponse transofrm response body
		TransformResponse []TransformResponse
		// Headers  custom headers for request
		Headers http.Header
		// Params params for request route
		Params map[string]string
		// Query query for requset
		Query url.Values

		// Body the request body
		Body interface{}

		// Concurrency current amount handling request of instance
		Concurrency uint32

		// Timeout request timeout
		Timeout time.Duration

		// Context context
		Context context.Context

		// Client http client
		Client *http.Client
		// Adapter custom handling of requset
		Adapter Adapter
		// RequestInterceptors request interceptor list
		RequestInterceptors []RequestInterceptor
		// ResponseInterceptors response interceptor list
		ResponseInterceptors []ResponseInterceptor

		// OnError on error function
		OnError OnError

		HTTPTrace   *HT.HTTPTrace
		enableTrace bool
		data        map[string]interface{}
	}
	// InstanceConfig config of instance
	InstanceConfig struct {
		// BaseURL http request base url
		BaseURL string
		// TransformRequest transform requset body
		TransformRequest []TransformRequest
		// TransformResponse transofrm response body
		TransformResponse []TransformResponse
		// Headers  custom headers for request
		Headers http.Header
		// Timeout request timeout
		Timeout time.Duration

		// Client http client
		Client *http.Client
		// Adapter custom adapter
		Adapter Adapter

		// RequestInterceptors request interceptor list
		RequestInterceptors []RequestInterceptor
		// ResponseInterceptors response interceptor list
		ResponseInterceptors []ResponseInterceptor

		// EnableTrace enable http trace
		EnableTrace bool
		// OnError on error function
		OnError OnError
	}
)

// Get get value from config
func (conf *Config) Get(key string) interface{} {
	if conf.data == nil {
		return nil
	}
	return conf.data[key]
}

// Set set value to config
func (conf *Config) Set(key string, value interface{}) {
	if conf.data == nil {
		conf.data = make(map[string]interface{})
	}
	conf.data[key] = value
}

// GetString get string value
func (conf *Config) GetString(key string) string {
	return cast.ToString(conf.Get(key))
}

// GetBool get bool value
func (conf *Config) GetBool(key string) bool {
	return cast.ToBool(conf.Get(key))
}

// GetInt get int value
func (conf *Config) GetInt(key string) int {
	return cast.ToInt(conf.Get(key))
}
