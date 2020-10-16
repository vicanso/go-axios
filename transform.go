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
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/andybalholm/brotli"
)

type (
	// TransformResponse transform function for http response
	TransformResponse func(body []byte, headers http.Header) (data []byte, err error)
	// TransformRequest transform function for http request
	TransformRequest func(body interface{}, headers http.Header) (data interface{}, err error)

	newReader func(io.Reader) (io.Reader, error)
)

func createEncodingTransform(encoding string, fn newReader) TransformResponse {
	return func(body []byte, headers http.Header) (data []byte, err error) {
		// 判断数据的encoding，如果不相等，直接返回
		if !strings.EqualFold(headers.Get(headerContentEncoding), encoding) {
			return body, nil
		}
		r := bytes.NewReader(body)
		reader, err := fn(r)
		if err != nil {
			return
		}
		data, err = ioutil.ReadAll(reader)
		if err != nil {
			return
		}
		// 完成后删除encoding以及content-length
		headers.Del(headerContentEncoding)
		headers.Del(headerContentLength)
		return
	}
}

func newGzipReader(r io.Reader) (io.Reader, error) {
	return gzip.NewReader(r)
}

func newBrReader(r io.Reader) (io.Reader, error) {
	return brotli.NewReader(r), nil
}

func setContentTypeIfUnset(headers http.Header, value string) {
	if headers.Get(headerContentType) == "" {
		headers.Set(headerContentType, value)
	}
}

func convertRequestBody(data interface{}, headers http.Header) (body interface{}, err error) {
	switch data := data.(type) {
	case []byte:
		body = data
	case string:
		body = []byte(data)
	case url.Values:
		v := data
		body = []byte(v.Encode())
		setContentTypeIfUnset(headers, contentTypeWWWFormUrlencoded)
	default:
		body, err = jsonMarshal(data)
		// 如果成功转换
		if err == nil {
			setContentTypeIfUnset(headers, contentTypeJSON)
		}
	}
	return
}

var (
	// DefaultTransformResponse default transform response
	DefaultTransformResponse []TransformResponse
	// DefaultTransformRequest default transform request
	DefaultTransformRequest []TransformRequest
)

func init() {
	DefaultTransformResponse = []TransformResponse{
		createEncodingTransform(gzipEncoding, newGzipReader),
		createEncodingTransform(brEncoding, newBrReader),
	}
	DefaultTransformRequest = []TransformRequest{
		convertRequestBody,
	}
}
