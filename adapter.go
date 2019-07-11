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
	"io/ioutil"
	"net/http"
)

type (
	// Adapter adapter function
	Adapter func(config *Config) (resp *Response, err error)
)

func defaultAdapter(config *Config) (resp *Response, err error) {

	req := config.Request

	// 发送请求
	client := config.Client
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}

	resp = &Response{
		Status:  res.StatusCode,
		Headers: res.Header,
	}
	// 读取数据
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	resp.Data = data

	return
}
