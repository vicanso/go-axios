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
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// RetryDecide decides whether another attempt should be made after a failed try.
// attempt is 1-based (1 = first try just finished).
type RetryDecide func(ctx context.Context, config *Config, resp *Response, err error, attempt int) bool

// RetryConfig is the retry policy for a client or a single request.
//
// Total attempts = MaxRetries + 1. Zero MaxRetries disables retry.
//
// Request bodies that implement io.Reader (other than values re-encoded each
// attempt such as structs/maps) are not replayable; retries are skipped for
// those requests. Prefer []byte, string, url.Values, or a value that is
// JSON-marshaled on each attempt.
type RetryConfig struct {
	// MaxRetries is the number of retries after the first attempt.
	MaxRetries int
	// MinWait is the base delay before the first retry. Default 100ms when <= 0.
	MinWait time.Duration
	// MaxWait caps the exponential backoff delay. Zero means no cap.
	MaxWait time.Duration
	// Multiplier is the exponential backoff factor. Default 2 when < 1.
	Multiplier float64
	// Methods lists HTTP methods that may be retried.
	// Empty means GET, HEAD, and OPTIONS only.
	Methods []string
	// Decide overrides DefaultRetryIf when non-nil.
	Decide RetryDecide
}

// DefaultRetryConfig returns a conservative retry policy:
// 2 retries, 100ms base wait, 2s max wait, exponential multiplier 2,
// idempotent methods only, DefaultRetryIf.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 2,
		MinWait:    100 * time.Millisecond,
		MaxWait:    2 * time.Second,
		Multiplier: 2,
	}
}

// DefaultRetryIf retries transport failures (timeout, DNS, connection reset, etc.)
// and selected HTTP status codes (408, 429, 500, 502, 503, 504).
// It does not retry context cancellation, body-too-large, or other status errors.
func DefaultRetryIf(ctx context.Context, config *Config, resp *Response, err error, attempt int) bool {
	if err == nil {
		return false
	}
	if ctx.Err() != nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, ErrBodyTooLarge) {
		return false
	}
	if errors.Is(err, ErrRequestIsForbidden) || errors.Is(err, ErrTooManyRequests) {
		return false
	}
	if errors.Is(err, ErrStatusInvalid) {
		if resp == nil {
			return false
		}
		switch resp.Status {
		case http.StatusRequestTimeout, // 408
			http.StatusTooManyRequests,     // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout:      // 504
			return true
		}
		return false
	}

	switch GetInternalErrorCategory(err) {
	case ErrCategoryTimeout,
		ErrCategoryDNS,
		ErrCategoryRefused,
		ErrCategoryReset,
		ErrCategoryAborted:
		return true
	}
	return false
}

func (cfg *RetryConfig) maxAttempts() int {
	if cfg == nil || cfg.MaxRetries <= 0 {
		return 1
	}
	return cfg.MaxRetries + 1
}

func (cfg *RetryConfig) methodAllowed(method string) bool {
	if cfg == nil {
		return false
	}
	method = strings.ToUpper(method)
	if method == "" {
		method = http.MethodGet
	}
	methods := cfg.Methods
	if len(methods) == 0 {
		switch method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			return true
		default:
			return false
		}
	}
	for _, m := range methods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

func (cfg *RetryConfig) decide() RetryDecide {
	if cfg != nil && cfg.Decide != nil {
		return cfg.Decide
	}
	return DefaultRetryIf
}

func (cfg *RetryConfig) delay(retryIndex int) time.Duration {
	// retryIndex is 0 for the first retry (after attempt 1).
	minWait := 100 * time.Millisecond
	if cfg != nil && cfg.MinWait > 0 {
		minWait = cfg.MinWait
	}
	mult := 2.0
	if cfg != nil && cfg.Multiplier >= 1 {
		mult = cfg.Multiplier
	}
	d := float64(minWait) * math.Pow(mult, float64(retryIndex))
	wait := time.Duration(d)
	if cfg != nil && cfg.MaxWait > 0 && wait > cfg.MaxWait {
		wait = cfg.MaxWait
	}
	if wait < 0 {
		return 0
	}
	return wait
}

// isReplayableBody reports whether Body can safely be sent again.
// io.Reader values are consumed on the first attempt and are not replayable.
func isReplayableBody(body any) bool {
	if body == nil {
		return true
	}
	if _, ok := body.(io.Reader); ok {
		return false
	}
	return true
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func shouldRetry(ctx context.Context, config *Config, resp *Response, err error, attempt int) bool {
	if err == nil {
		return false
	}
	cfg := config.Retry
	if cfg == nil || attempt >= cfg.maxAttempts() {
		return false
	}
	if !cfg.methodAllowed(config.Method) {
		return false
	}
	if !isReplayableBody(config.Body) {
		return false
	}
	return cfg.decide()(ctx, config, resp, err, attempt)
}
