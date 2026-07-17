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
	"context"
	"encoding/json"
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

// newHTTPClient creates a http client which will not be
// intercepted by gock
func newHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{},
	}
}

// newGockClient creates a client intercepted by gock.
func newGockClient(opts ...ClientOption) *Client {
	c := NewClient(opts...)
	gock.InterceptClient(c.GetConfig().HTTPClient)
	return c
}

func TestClientGet(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/users/me").
		MatchParam("t", "1").
		Reply(200).
		JSON(map[string]string{
			"account": "tree",
		})

	c := newGockClient(
		WithBaseURL("http://aslant.site"),
	)
	resp, err := c.Get(context.Background(), "/users/me", Query("t", "1"))
	assert.Nil(err)
	assert.Equal(200, resp.Status)
	assert.True(resp.OK())

	user := userInfo{}
	err = resp.JSON(&user)
	assert.Nil(err)
	assert.Equal("tree", user.Account)
}

func TestClientParam(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/users/vip").
		Reply(200).
		BodyString("ok")

	c := newGockClient(WithBaseURL("http://aslant.site"))
	resp, err := c.Get(context.Background(), "/users/:type", Param("type", "vip"))
	assert.Nil(err)
	assert.Equal(200, resp.Status)
	assert.Equal("ok", string(resp.Data))
	// route为原始的url
	assert.Equal("/users/:type", resp.Config.Route)
}

func TestClientPostJSONBody(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set(headerContentType, r.Header.Get(headerContentType))
		w.WriteHeader(201)
		_, _ = w.Write(body)
	}))
	defer ts.Close()

	c := NewClient(WithHTTPClient(newHTTPClient()))
	resp, err := c.Post(context.Background(), ts.URL+"/users", map[string]string{
		"account": "tree",
	})
	assert.Nil(err)
	assert.Equal(201, resp.Status)
	assert.True(resp.OK())
	assert.Equal(`{"account":"tree"}`, string(resp.Data))
	assert.Equal(contentTypeJSON, resp.Headers.Get(headerContentType))
}

func TestClientDefaultHeaders(t *testing.T) {
	assert := assert.New(t)

	var request *http.Request
	c := NewClient(
		WithHeader("X-Client", "go-axios"),
		WithAdapter(func(config *Config) (*Response, error) {
			request = config.Request
			return &Response{Status: 200}, nil
		}),
	)
	_, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Nil(err)
	assert.Equal(UserAgent, request.UserAgent())
	assert.Equal(defaultAcceptEncoding, request.Header.Get(headerAcceptEncoding))
	assert.Equal(defaultAccept, request.Header.Get(headerAccept))
	assert.Equal("go-axios", request.Header.Get("X-Client"))
}

func TestClientInterceptorOrder(t *testing.T) {
	assert := assert.New(t)

	orders := make([]string, 0)
	c := NewClient(
		WithOnBeforeRequest(func(config *Config) error {
			orders = append(orders, "before")
			return nil
		}),
		WithRequestInterceptor(func(config *Config) error {
			orders = append(orders, "request")
			return nil
		}),
		WithResponseInterceptor(func(resp *Response) error {
			orders = append(orders, "response")
			return nil
		}),
		WithOnDone(func(config *Config, resp *Response, err error) {
			orders = append(orders, "done")
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			orders = append(orders, "adapter")
			return &Response{Status: 200}, nil
		}),
	)
	resp, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Nil(err)
	assert.Equal(200, resp.Status)
	assert.Equal([]string{"before", "request", "adapter", "response", "done"}, orders)
}

func TestClientOnBeforeRequestShortCircuit(t *testing.T) {
	assert := assert.New(t)

	c := NewClient(
		WithOnBeforeRequest(func(config *Config) error {
			config.Response = &Response{
				Status: 200,
				Data:   []byte("from cache"),
			}
			return nil
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			return nil, errors.New("should not be called")
		}),
	)
	resp, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Nil(err)
	assert.Equal("from cache", string(resp.Data))
	assert.NotNil(resp.Config)
}

func TestClientOnError(t *testing.T) {
	assert := assert.New(t)

	customErr := errors.New("custom error")
	var doneErr error
	c := NewClient(
		WithAdapter(func(config *Config) (*Response, error) {
			return nil, errors.New("original error")
		}),
		WithOnError(func(err error, config *Config) error {
			return customErr
		}),
		WithOnDone(func(config *Config, resp *Response, err error) {
			doneErr = err
		}),
	)
	_, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Equal(customErr, err)
	assert.Equal(customErr, doneErr)
}

func TestClientNilResponse(t *testing.T) {
	assert := assert.New(t)

	c := NewClient(
		WithAdapter(func(config *Config) (*Response, error) {
			return nil, nil
		}),
	)
	_, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Equal(ErrNilResponse, err)
}

func TestClientConcurrencyLimit(t *testing.T) {
	assert := assert.New(t)

	c := NewClient()
	c.SetMaxConcurrency(-1)
	_, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Equal(ErrRequestIsForbidden, err)

	c.SetMaxConcurrency(1)
	block := make(chan struct{})
	done := c.CustomMock(func(config *Config) (*Response, error) {
		<-block
		return &Response{Status: 200}, nil
	})
	defer done()

	firstDone := make(chan error, 1)
	go func() {
		_, err := c.Get(context.Background(), "http://aslant.site/")
		firstDone <- err
	}()
	// 等待第一个请求进入adapter
	for range 100 {
		if c.GetConcurrency() == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(int32(1), c.GetConcurrency())

	_, err = c.Get(context.Background(), "http://aslant.site/")
	assert.Equal(ErrTooManyRequests, err)

	close(block)
	assert.Nil(<-firstDone)
	assert.Equal(int32(0), c.GetConcurrency())
}

func TestClientMock(t *testing.T) {
	assert := assert.New(t)

	c := NewClient()
	done := c.Mock(&Response{
		Status: 429,
		Data:   []byte("mock"),
	})
	resp, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Nil(err)
	assert.Equal(429, resp.Status)
	assert.Equal("mock", string(resp.Data))
	done()
}

func TestClientMultiMock(t *testing.T) {
	assert := assert.New(t)

	c := NewClient(WithBaseURL("http://aslant.site"))
	done := c.MultiMock(map[string]*Response{
		"/users/me": {
			Status: 200,
			Data:   []byte("me"),
		},
		"/books": {
			Status: 200,
			Data:   []byte("books"),
		},
	})
	defer done()

	resp, err := c.Get(context.Background(), "/users/me")
	assert.Nil(err)
	assert.Equal("me", string(resp.Data))

	resp, err = c.Get(context.Background(), "/books")
	assert.Nil(err)
	assert.Equal("books", string(resp.Data))

	// 未匹配的route返回nil response
	_, err = c.Get(context.Background(), "/not-found")
	assert.Equal(ErrNilResponse, err)
}

func TestClientTimeout(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c := NewClient(
		WithHTTPClient(newHTTPClient()),
		WithTimeout(20*time.Millisecond),
	)
	_, err := c.Get(context.Background(), ts.URL)
	assert.NotNil(err)
	assert.Equal(ErrCategoryTimeout, GetInternalErrorCategory(err))
}

func TestClientEnableTrace(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	c := NewClient(
		WithHTTPClient(newHTTPClient()),
		WithEnableTrace(),
	)
	resp, err := c.Get(context.Background(), ts.URL)
	assert.Nil(err)
	assert.NotNil(resp.Config.HTTPTrace)

	stats := GetStats(resp.Config, nil)
	assert.Equal(200, stats.Status)
	assert.Equal(ResultSuccess, stats.Result)
	assert.Equal(2, stats.Size)
}

func TestClientGzipResponse(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerContentEncoding, gzipEncoding)
		gz := gzip.NewWriter(w)
		_, _ = gz.Write([]byte(`{"account":"tree"}`))
		_ = gz.Close()
	}))
	defer ts.Close()

	c := NewClient(WithHTTPClient(newHTTPClient()))
	resp, err := c.Get(context.Background(), ts.URL)
	assert.Nil(err)
	assert.Equal(`{"account":"tree"}`, string(resp.Data))
	assert.Empty(resp.Headers.Get(headerContentEncoding))
}

func TestClientUpload(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(1024 * 1024)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(400)
			return
		}
		buf, _ := io.ReadAll(file)
		_, _ = w.Write([]byte(r.FormValue("type") + ":" + string(buf)))
	}))
	defer ts.Close()

	form := NewMultipartForm()
	err := form.AddFile("file", "test.txt", []byte("hello"))
	assert.Nil(err)
	err = form.AddField("type", "vip")
	assert.Nil(err)

	c := NewClient(WithHTTPClient(newHTTPClient()))
	resp, err := c.Upload(context.Background(), ts.URL, form)
	assert.Nil(err)
	assert.Equal("vip:hello", string(resp.Data))
}

func TestClientCustomJSON(t *testing.T) {
	assert := assert.New(t)

	marshalCount := 0
	unmarshalCount := 0
	c := NewClient(
		WithJSONMarshal(func(v any) ([]byte, error) {
			marshalCount++
			return json.Marshal(v)
		}),
		WithJSONUnmarshal(func(data []byte, v any) error {
			unmarshalCount++
			return json.Unmarshal(data, v)
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			return &Response{
				Status: 200,
				Data:   []byte(`{"account":"tree"}`),
			}, nil
		}),
	)
	user, err := PostJSON[userInfo](context.Background(), c, "http://aslant.site/", map[string]string{
		"account": "tree",
	})
	assert.Nil(err)
	assert.Equal("tree", user.Account)
	assert.Equal(1, marshalCount)
	assert.Equal(1, unmarshalCount)
}

func TestClientRequestInterceptorError(t *testing.T) {
	assert := assert.New(t)

	customErr := errors.New("interceptor error")
	c := NewClient(
		WithRequestInterceptor(func(config *Config) error {
			return customErr
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			return &Response{Status: 200}, nil
		}),
	)
	_, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Equal(customErr, err)
}

func TestClientResponseInterceptorError(t *testing.T) {
	assert := assert.New(t)

	customErr := errors.New("response interceptor error")
	c := NewClient(
		WithResponseInterceptor(func(resp *Response) error {
			return customErr
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			return &Response{Status: 200}, nil
		}),
	)
	_, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Equal(customErr, err)
}

func TestClientContextCancel(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	c := NewClient(WithHTTPClient(newHTTPClient()))
	_, err := c.Get(ctx, ts.URL)
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), "context canceled"))
}

func TestClientGetConfig(t *testing.T) {
	assert := assert.New(t)

	c := NewClient(
		WithBaseURL("http://aslant.site"),
		WithTimeout(time.Second),
		WithMaxConcurrency(10),
		WithHeader("X-Token", "secret"),
		WithHeaders(http.Header{"X-Extra": {"1"}}),
		WithTransformRequest(DefaultTransformRequest...),
		WithTransformResponse(DefaultTransformResponse...),
		WithRequestInterceptor(func(config *Config) error { return nil }),
		WithResponseInterceptor(func(resp *Response) error { return nil }),
		WithOnError(func(err error, config *Config) error { return nil }),
		WithOnDone(func(config *Config, resp *Response, err error) {}),
		WithOnBeforeRequest(func(config *Config) error { return nil }),
		WithMaxBodySize(1024),
		WithValidateStatus(ValidateStatus2xx),
	)
	config := c.GetConfig()
	assert.Equal("http://aslant.site", config.BaseURL)
	assert.Equal(time.Second, config.Timeout)
	assert.Equal(int32(10), config.MaxConcurrency)
	assert.Equal(int64(1024), config.MaxBodySize)
	assert.NotNil(config.ValidateStatus)
	assert.NotNil(config.HTTPClient)
	assert.Equal("1", config.Headers.Get("X-Extra"))

	// GetConfig returns a deep copy of headers and slices
	config.Headers.Set("X-Token", "mutated")
	config.RequestInterceptors = nil
	assert.Equal("secret", c.GetConfig().Headers.Get("X-Token"))
	assert.Len(c.GetConfig().RequestInterceptors, 1)
}

func TestClientMaxBodySize(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("0123456789"))
	}))
	defer ts.Close()

	c := NewClient(
		WithHTTPClient(newHTTPClient()),
		WithMaxBodySize(5),
	)
	resp, err := c.Get(context.Background(), ts.URL)
	assert.NotNil(err)
	assert.True(errors.Is(err, ErrBodyTooLarge))
	var axErr *Error
	assert.True(errors.As(err, &axErr))
	assert.Equal(ErrBodyTooLarge, axErr.Err)
	assert.Equal(5, len(resp.Data))

	// request option overrides unlimited client
	c = NewClient(WithHTTPClient(newHTTPClient()))
	resp, err = c.Get(context.Background(), ts.URL, MaxBodySize(3))
	assert.True(errors.Is(err, ErrBodyTooLarge))
	assert.Equal(3, len(resp.Data))
}

func TestClientValidateStatus(t *testing.T) {
	assert := assert.New(t)

	c := NewClient(
		WithValidateStatus(ValidateStatus2xx),
		WithAdapter(func(config *Config) (*Response, error) {
			return &Response{
				Status: 404,
				Data:   []byte("missing"),
			}, nil
		}),
	)
	resp, err := c.Get(context.Background(), "http://aslant.site/")
	assert.NotNil(err)
	assert.NotNil(resp)
	assert.Equal(404, resp.Status)
	assert.True(errors.Is(err, ErrStatusInvalid))
	var axErr *Error
	assert.True(errors.As(err, &axErr))
	assert.Equal(404, axErr.Status)
	assert.Equal("missing", string(axErr.Response.Data))

	// request option
	c = NewClient(
		WithAdapter(func(config *Config) (*Response, error) {
			return &Response{Status: 500, Data: []byte("err")}, nil
		}),
	)
	_, err = c.Get(context.Background(), "http://aslant.site/", StatusValidator(ValidateStatus2xx))
	assert.True(errors.Is(err, ErrStatusInvalid))

	// 2xx still succeeds
	c = NewClient(
		WithValidateStatus(ValidateStatus2xx),
		WithAdapter(func(config *Config) (*Response, error) {
			return &Response{Status: 201, Data: []byte("created")}, nil
		}),
	)
	resp, err = c.Get(context.Background(), "http://aslant.site/")
	assert.Nil(err)
	assert.Equal(201, resp.Status)
}

func TestClientDefaultHTTPClientIsolated(t *testing.T) {
	assert := assert.New(t)

	c1 := NewClient()
	c2 := NewClient()
	assert.NotNil(c1.GetConfig().HTTPClient)
	assert.NotNil(c2.GetConfig().HTTPClient)
	assert.NotSame(c1.GetConfig().HTTPClient, c2.GetConfig().HTTPClient)
	assert.NotSame(c1.GetConfig().HTTPClient, http.DefaultClient)
}
