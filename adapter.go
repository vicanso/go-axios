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
	"io/ioutil"
	"net/http"
	"net/http/httptrace"

	HT "github.com/vicanso/http-trace"
)

type (
	// Adapter adapter function
	Adapter func(config *Config) (resp *Response, err error)
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

func defaultAdapter(config *Config) (resp *Response, err error) {
	var r io.Reader
	if config.Body != nil && isNeedToTransformRequestBody(config.Method) {
		data := config.Body
		for _, fn := range config.TransformRequest {
			buf, e := fn(data, config.Headers)
			if e != nil {
				err = CreateError(e, config, 0)
				return
			}
			data = buf
		}
		r = bytes.NewReader(data.([]byte))
	}

	req, err := http.NewRequest(config.Method, config.URL, r)
	if err != nil {
		err = CreateError(err, config, 0)
		return
	}
	config.Request = req

	if config.enableTrace {
		ctx := config.Context
		if ctx == nil {
			ctx = context.Background()
		}
		trace, ht := HT.NewClientTrace()
		ctx = httptrace.WithClientTrace(ctx, trace)
		req.WithContext(ctx)
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

	// 请求前的相关拦截器调用
	for _, fn := range config.RequestInterceptors {
		err = fn(config)
		if err != nil {
			err = CreateError(err, config, 0)
			return
		}
	}

	// 发送请求
	client := config.Client
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		err = CreateError(err, config, 0)
		return
	}

	resp = &Response{
		Status:  res.StatusCode,
		Headers: res.Header,
		Config:  config,
		Request: req,
	}
	config.Response = resp
	// 读取数据
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = CreateError(err, config, resp.Status)
		return
	}

	// 响应数据的相关转换
	for _, fn := range config.TransformResponse {
		data, err = fn(data, resp.Headers)
		if err != nil {
			err = CreateError(err, config, http.StatusInternalServerError)
			return
		}
	}
	resp.Data = data

	// 响应完成后的相关响应拦截器
	for _, fn := range config.ResponseInterceptors {
		err = fn(resp)
		if err != nil {
			err = CreateError(err, config, http.StatusInternalServerError)
			return
		}
	}
	return
}
