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
	"io"
	"net/http"
	"strconv"
)

type (
	// Adapter adapter function
	Adapter func(config *Config) (resp *Response, err error)
)

// copy from io.ReadAll
// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.
func ReadAllInitCap(r io.Reader, initCap int) ([]byte, error) {
	if initCap <= 0 {
		initCap = 512
	}
	b := make([]byte, 0, initCap)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}

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
		Status:           res.StatusCode,
		Headers:          res.Header,
		OriginalResponse: res,
	}
	// 读取数据
	defer res.Body.Close()
	size := 0
	contentLength := res.Header.Get("Content-Length")
	if contentLength != "" {
		size, _ = strconv.Atoi(contentLength)
	}
	data, err := ReadAllInitCap(res.Body, size)
	if err != nil {
		return
	}
	resp.Data = data

	return
}
