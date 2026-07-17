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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestAdapter(t *testing.T) {
	t.Run("request get", func(t *testing.T) {
		assert := assert.New(t)
		defer gock.Off()
		data := "abcd"
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			BodyString(data)

		config := &Config{
			URL:     "http://aslant.site/",
			Method:  "GET",
			Timeout: time.Second,
		}
		config.Request, _ = newRequest(config)
		resp, err := defaultAdapter(config)
		assert.Nil(err)
		assert.Equal(data, string(resp.Data))
	})

	t.Run("request error", func(t *testing.T) {
		assert := assert.New(t)
		defer gock.Off()
		data := "abcd"
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			BodyString(data)
		config := &Config{
			URL:     "http://aslant.site/",
			Method:  "POST",
			Timeout: time.Second,
		}
		config.Request, _ = newRequest(config)
		_, err := defaultAdapter(config)
		assert.NotNil(err)
		assert.True(strings.Contains(err.Error(), "cannot match any request"))
	})
}

func TestAdapterMaxBodySize(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world"))
	}))
	defer ts.Close()

	config := &Config{
		URL:         ts.URL,
		Method:      http.MethodGet,
		MaxBodySize: 5,
		HTTPClient:  &http.Client{Transport: &http.Transport{}},
	}
	config.Request, _ = newRequest(config)
	resp, err := defaultAdapter(config)
	assert.True(errors.Is(err, ErrBodyTooLarge))
	assert.Equal("hello", string(resp.Data))
}

func TestAdapterDecodeGzipWhileReading(t *testing.T) {
	assert := assert.New(t)

	raw := []byte("decoded-body")
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(raw)
	_ = zw.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentEncoding, gzipEncoding)
		_, _ = w.Write(buf.Bytes())
	}))
	defer ts.Close()

	config := &Config{
		URL:        ts.URL,
		Method:     http.MethodGet,
		HTTPClient: &http.Client{Transport: &http.Transport{}},
	}
	config.Request, _ = newRequest(config)
	resp, err := defaultAdapter(config)
	assert.Nil(err)
	assert.Equal(string(raw), string(resp.Data))
	assert.Empty(resp.Headers.Get(headerContentEncoding))
}

func TestReadAllInitCap(t *testing.T) {
	assert := assert.New(t)
	data, err := ReadAllInitCap(strings.NewReader("abc"), 0)
	assert.Nil(err)
	assert.Equal("abc", string(data))

	data, err = ReadAllInitCap(io.NopCloser(strings.NewReader("xyz")), 2)
	assert.Nil(err)
	assert.Equal("xyz", string(data))
}

func TestWrapDecodingReader(t *testing.T) {
	assert := assert.New(t)

	// no encoding
	h := make(http.Header)
	r, c, err := wrapDecodingReader(strings.NewReader("plain"), h)
	assert.Nil(err)
	assert.Nil(c)
	data, _ := io.ReadAll(r)
	assert.Equal("plain", string(data))

	// unknown encoding left as-is
	h = make(http.Header)
	h.Set(headerContentEncoding, "deflate")
	r, c, err = wrapDecodingReader(strings.NewReader("x"), h)
	assert.Nil(err)
	assert.Nil(c)
	assert.Equal("deflate", h.Get(headerContentEncoding))
	data, _ = io.ReadAll(r)
	assert.Equal("x", string(data))

	// invalid gzip
	h = make(http.Header)
	h.Set(headerContentEncoding, gzipEncoding)
	_, _, err = wrapDecodingReader(strings.NewReader("not-gzip"), h)
	assert.NotNil(err)
}
