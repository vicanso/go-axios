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
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRetryIf(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	cfg := &Config{}

	assert.False(DefaultRetryIf(ctx, cfg, nil, nil, 1))
	assert.False(DefaultRetryIf(ctx, cfg, nil, context.Canceled, 1))
	assert.False(DefaultRetryIf(ctx, cfg, &Response{Status: 400}, ErrBodyTooLarge, 1))
	assert.False(DefaultRetryIf(ctx, cfg, &Response{Status: 404}, newStatusError(cfg, &Response{Status: 404}), 1))
	assert.True(DefaultRetryIf(ctx, cfg, &Response{Status: 503}, newStatusError(cfg, &Response{Status: 503}), 1))
	assert.True(DefaultRetryIf(ctx, cfg, nil, &timeoutErr{}, 1))

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	assert.False(DefaultRetryIf(canceled, cfg, nil, errors.New("x"), 1))
}

func TestRetryConfigHelpers(t *testing.T) {
	assert := assert.New(t)

	var nilCfg *RetryConfig
	assert.Equal(1, nilCfg.maxAttempts())
	assert.False(nilCfg.methodAllowed(http.MethodGet))

	cfg := DefaultRetryConfig()
	assert.Equal(3, cfg.maxAttempts())
	assert.True(cfg.methodAllowed(http.MethodGet))
	assert.True(cfg.methodAllowed(http.MethodHead))
	assert.False(cfg.methodAllowed(http.MethodPost))
	assert.Equal(100*time.Millisecond, cfg.delay(0))
	assert.Equal(200*time.Millisecond, cfg.delay(1))
	assert.Equal(400*time.Millisecond, cfg.delay(2))

	cfg.MaxWait = 250 * time.Millisecond
	assert.Equal(250*time.Millisecond, cfg.delay(2))

	cfg.Methods = []string{"POST"}
	assert.True(cfg.methodAllowed(http.MethodPost))
	assert.False(cfg.methodAllowed(http.MethodGet))
}

func TestIsReplayableBody(t *testing.T) {
	assert := assert.New(t)
	assert.True(isReplayableBody(nil))
	assert.True(isReplayableBody([]byte("a")))
	assert.True(isReplayableBody("a"))
	assert.True(isReplayableBody(map[string]string{"a": "1"}))
	assert.False(isReplayableBody(bytes.NewBufferString("x")))
	assert.False(isReplayableBody(bytes.NewReader([]byte("x"))))
}

func TestClientRetryOnStatus(t *testing.T) {
	assert := assert.New(t)

	var calls atomic.Int32
	c := NewClient(
		WithValidateStatus(ValidateStatus2xx),
		WithRetry(RetryConfig{
			MaxRetries: 2,
			MinWait:    time.Millisecond,
			MaxWait:    time.Millisecond,
			Multiplier: 1,
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			n := calls.Add(1)
			if n < 3 {
				return &Response{Status: 503, Data: []byte("busy")}, nil
			}
			return &Response{Status: 200, Data: []byte("ok")}, nil
		}),
	)

	resp, err := c.Get(context.Background(), "http://aslant.site/")
	assert.Nil(err)
	assert.Equal(200, resp.Status)
	assert.Equal("ok", string(resp.Data))
	assert.Equal(int32(3), calls.Load())
	assert.Equal(3, resp.Config.Attempt)
}

func TestClientRetryExhausted(t *testing.T) {
	assert := assert.New(t)

	var calls atomic.Int32
	c := NewClient(
		WithValidateStatus(ValidateStatus2xx),
		WithRetry(RetryConfig{
			MaxRetries: 1,
			MinWait:    time.Millisecond,
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			calls.Add(1)
			return &Response{Status: 503, Data: []byte("busy")}, nil
		}),
	)
	resp, err := c.Get(context.Background(), "http://aslant.site/")
	assert.True(errors.Is(err, ErrStatusInvalid))
	assert.Equal(503, resp.Status)
	assert.Equal(int32(2), calls.Load())
	assert.Equal(2, resp.Config.Attempt)
}

func TestClientRetrySkipsNonIdempotent(t *testing.T) {
	assert := assert.New(t)

	var calls atomic.Int32
	c := NewClient(
		WithValidateStatus(ValidateStatus2xx),
		WithRetry(DefaultRetryConfig()),
		WithAdapter(func(config *Config) (*Response, error) {
			calls.Add(1)
			return &Response{Status: 503}, nil
		}),
	)
	_, err := c.Post(context.Background(), "http://aslant.site/", map[string]string{"a": "1"})
	assert.True(errors.Is(err, ErrStatusInvalid))
	assert.Equal(int32(1), calls.Load())
}

func TestClientRetrySkipsNonReplayableBody(t *testing.T) {
	assert := assert.New(t)

	var calls atomic.Int32
	c := NewClient(
		WithValidateStatus(ValidateStatus2xx),
		WithRetry(RetryConfig{
			MaxRetries: 2,
			MinWait:    time.Millisecond,
			Methods:    []string{http.MethodPost},
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			calls.Add(1)
			return &Response{Status: 503}, nil
		}),
	)
	_, err := c.Post(context.Background(), "http://aslant.site/", bytes.NewBufferString("x"))
	assert.True(errors.Is(err, ErrStatusInvalid))
	assert.Equal(int32(1), calls.Load())
}

func TestClientRetryRequestOption(t *testing.T) {
	assert := assert.New(t)

	var calls atomic.Int32
	c := NewClient(
		WithValidateStatus(ValidateStatus2xx),
		WithAdapter(func(config *Config) (*Response, error) {
			n := calls.Add(1)
			if n == 1 {
				return &Response{Status: 503}, nil
			}
			return &Response{Status: 200, Data: []byte("ok")}, nil
		}),
	)
	resp, err := c.Get(context.Background(), "http://aslant.site/",
		RetryPolicy(RetryConfig{MaxRetries: 1, MinWait: time.Millisecond}),
	)
	assert.Nil(err)
	assert.Equal("ok", string(resp.Data))
	assert.Equal(int32(2), calls.Load())
}

func TestClientRetryCustomDecide(t *testing.T) {
	assert := assert.New(t)

	var calls atomic.Int32
	c := NewClient(
		WithRetry(RetryConfig{
			MaxRetries: 3,
			MinWait:    time.Millisecond,
			Decide: func(ctx context.Context, config *Config, resp *Response, err error, attempt int) bool {
				return false
			},
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			calls.Add(1)
			return nil, errors.New("always fail")
		}),
	)
	_, err := c.Get(context.Background(), "http://aslant.site/")
	assert.EqualError(err, "always fail")
	assert.Equal(int32(1), calls.Load())
}

func TestClientRetryRespectsContextCancel(t *testing.T) {
	assert := assert.New(t)

	var calls atomic.Int32
	c := NewClient(
		WithValidateStatus(ValidateStatus2xx),
		WithRetry(RetryConfig{
			MaxRetries: 5,
			MinWait:    50 * time.Millisecond,
		}),
		WithAdapter(func(config *Config) (*Response, error) {
			calls.Add(1)
			return &Response{Status: 503}, nil
		}),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := c.Get(ctx, "http://aslant.site/")
	assert.NotNil(err)
	// first attempt + maybe sleep canceled before more retries
	assert.LessOrEqual(calls.Load(), int32(2))
}

func TestClientClone(t *testing.T) {
	assert := assert.New(t)

	base := NewClient(
		WithBaseURL("http://aslant.site"),
		WithHeader("X-Base", "1"),
		WithTimeout(time.Second),
		WithRetry(DefaultRetryConfig()),
		WithValidateStatus(ValidateStatus2xx),
	)

	child := base.Clone(
		WithBaseURL("http://child.site"),
		WithHeader("X-Child", "2"),
	)

	baseCfg := base.GetConfig()
	childCfg := child.GetConfig()

	assert.Equal("http://aslant.site", baseCfg.BaseURL)
	assert.Equal("http://child.site", childCfg.BaseURL)
	assert.Equal("1", childCfg.Headers.Get("X-Base"))
	assert.Equal("2", childCfg.Headers.Get("X-Child"))
	assert.Equal("", baseCfg.Headers.Get("X-Child"))
	assert.Same(baseCfg.HTTPClient, childCfg.HTTPClient)
	assert.NotNil(childCfg.Retry)
	assert.Equal(2, childCfg.Retry.MaxRetries)

	// mutating clone headers does not affect base
	childCfg.Headers.Set("X-Base", "mutated")
	assert.Equal("1", base.GetConfig().Headers.Get("X-Base"))

	// mutating clone retry does not affect base
	childCfg.Retry.MaxRetries = 99
	assert.Equal(2, base.GetConfig().Retry.MaxRetries)

	// clone is independent concurrency counter
	assert.Equal(int32(0), child.GetConcurrency())
}
