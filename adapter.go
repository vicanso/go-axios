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
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

type (
	// Adapter adapter function
	Adapter func(config *Config) (resp *Response, err error)
)

// copy from io.ReadAll
// ReadAllInitCap reads from r until an error or EOF and returns the data it read,
// the buffer is initialized with the init capacity.
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

// wrapDecodingReader wraps body with a decoder based on Content-Encoding.
// On success it clears Content-Encoding and Content-Length so callers see
// decoded bytes. NeedClose is true when the decoder owns a Close method
// that must be called (gzip reader).
func wrapDecodingReader(body io.Reader, headers http.Header) (r io.Reader, needClose io.Closer, err error) {
	encoding := headers.Get(headerContentEncoding)
	if encoding == "" {
		return body, nil, nil
	}
	switch {
	case strings.EqualFold(encoding, gzipEncoding):
		gr, err := gzip.NewReader(body)
		if err != nil {
			return nil, nil, err
		}
		headers.Del(headerContentEncoding)
		headers.Del(headerContentLength)
		return gr, gr, nil
	case strings.EqualFold(encoding, brEncoding):
		headers.Del(headerContentEncoding)
		headers.Del(headerContentLength)
		return brotli.NewReader(body), nil, nil
	default:
		return body, nil, nil
	}
}

func defaultAdapter(config *Config) (resp *Response, err error) {
	req := config.Request

	// 发送请求
	client := config.HTTPClient
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
	defer func() {
		_ = res.Body.Close()
	}()

	body := io.Reader(res.Body)
	decoder, decoderCloser, err := wrapDecodingReader(body, res.Header)
	if err != nil {
		return
	}
	if decoderCloser != nil {
		defer func() {
			_ = decoderCloser.Close()
		}()
	}
	body = decoder

	// Limit is applied to decoded bytes so MaxBodySize matches payload size.
	if config.MaxBodySize > 0 {
		body = io.LimitReader(body, config.MaxBodySize+1)
	}

	size := 0
	// Content-Length is only a useful hint when the body is not decoded
	// (decoding clears Content-Encoding / Content-Length on the headers).
	contentLength := res.Header.Get(headerContentLength)
	if contentLength != "" {
		size, _ = strconv.Atoi(contentLength)
	}
	if config.MaxBodySize > 0 && size > 0 && int64(size) > config.MaxBodySize {
		size = int(config.MaxBodySize)
	}

	data, err := ReadAllInitCap(body, size)
	if err != nil {
		return
	}
	if config.MaxBodySize > 0 && int64(len(data)) > config.MaxBodySize {
		resp.Data = data[:config.MaxBodySize]
		return resp, newBodyTooLargeError(config, resp)
	}
	resp.Data = data

	return
}
