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
	"encoding/base64"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// doGzip gzip
func doGzip(buf []byte, level int) ([]byte, error) {
	var b bytes.Buffer
	if level <= 0 {
		level = gzip.DefaultCompression
	}
	w, _ := gzip.NewWriterLevel(&b, level)
	_, err := w.Write(buf)
	if err != nil {
		return nil, err
	}
	w.Close()
	return b.Bytes(), nil
}

func TestTransformResponse(t *testing.T) {
	t.Run("skip transform", func(t *testing.T) {
		assert := assert.New(t)
		buf := []byte("abcd")
		transform := createEncodingTransform(brEncoding, newBrReader)
		header := make(http.Header)
		data, err := transform(buf, header)
		assert.Nil(err)
		assert.Equal(buf, data)
	})

	t.Run("brotli transform", func(t *testing.T) {
		assert := assert.New(t)
		// abcd 做brotli压缩后的字符
		brDataB64 := "iwGAYWJjZAM="
		buf, err := base64.StdEncoding.DecodeString(brDataB64)
		assert.Nil(err)
		transform := createEncodingTransform(brEncoding, newBrReader)
		header := make(http.Header)
		header.Set(headerContentEncoding, brEncoding)
		header.Set(headerContentLength, "100")
		data, err := transform(buf, header)
		assert.Nil(err)
		assert.Equal("abcd", string(data))
		assert.Empty(header.Get(headerContentLength))
		assert.Empty(header.Get(headerContentEncoding))
	})

	t.Run("gzip transform", func(t *testing.T) {
		assert := assert.New(t)
		rawData := []byte("abcd")
		buf, err := doGzip(rawData, 0)
		assert.Nil(err)
		transform := createEncodingTransform(gzipEncoding, newGzipReader)
		header := make(http.Header)
		header.Set(headerContentEncoding, gzipEncoding)
		data, err := transform(buf, header)
		assert.Nil(err)
		assert.Equal(rawData, data)
		assert.Empty(header.Get(headerContentLength))
		assert.Empty(header.Get(headerContentEncoding))
	})
}

func TestConvertRequestBody(t *testing.T) {
	t.Run("www-form-urlencoded", func(t *testing.T) {
		assert := assert.New(t)
		data := make(url.Values)
		data.Add("a", "1")
		data.Add("a", "2")
		data.Add("b", "3")
		header := make(http.Header)
		body, err := convertRequestBody(data, header)
		assert.Nil(err)
		assert.Equal("a=1&a=2&b=3", string(body))
		assert.Equal(contentTypeWWWFormUrlencoded, header.Get(headerContentType))
	})

	t.Run("json", func(t *testing.T) {
		assert := assert.New(t)
		data := map[string]interface{}{
			"a": 1,
			"b": "2",
			"c": true,
		}
		header := make(http.Header)
		body, err := convertRequestBody(data, header)
		assert.Nil(err)
		assert.Equal(`{"a":1,"b":"2","c":true}`, string(body))
		assert.Equal(contentTypeJSON, header.Get(headerContentType))
	})

	t.Run("byte", func(t *testing.T) {
		assert := assert.New(t)
		data := []byte("abcd")
		header := make(http.Header)
		body, err := convertRequestBody(data, header)
		assert.Nil(err)
		assert.Equal(data, body)
		assert.Empty(header.Get(headerContentType))
	})
	t.Run("string", func(t *testing.T) {
		assert := assert.New(t)
		data := "abcd"
		header := make(http.Header)
		body, err := convertRequestBody(data, header)
		assert.Nil(err)
		assert.Equal([]byte(data), body)
		assert.Empty(header.Get(headerContentType))
	})
}
