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
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"syscall"
	"time"
)

const (
	ErrCategoryDNS      = "dns"
	ErrCategoryTimeout  = "timeout"
	ErrCategoryCanceled = "canceled"
	ErrCategoryAddr     = "addr"
	ErrCategoryAborted  = "aborted"
	ErrCategoryRefused  = "refused"
	ErrCategoryReset    = "reset"
)

const (
	ResultSuccess = iota
	ResultFail
)

type (
	// JSONMarshal json marshal function type
	JSONMarshal func(any) ([]byte, error)
	// JSONUnmarshal json unmarshal function type
	JSONUnmarshal func([]byte, any) error
)

var (
	defaultJSONMarshal   JSONMarshal   = json.Marshal
	defaultJSONUnmarshal JSONUnmarshal = json.Unmarshal
	// defaultClient is used by the package level request functions,
	// the timeout is set to one minute
	defaultClient = NewClient(WithTimeout(time.Minute))
)

func (conf *Config) marshalJSON(v any) ([]byte, error) {
	if conf.jsonMarshal != nil {
		return conf.jsonMarshal(v)
	}
	return defaultJSONMarshal(v)
}

func (conf *Config) unmarshalJSON(data []byte, v any) error {
	if conf.jsonUnmarshal != nil {
		return conf.jsonUnmarshal(data, v)
	}
	return defaultJSONUnmarshal(data, v)
}

// Default returns the default client
func Default() *Client {
	return defaultClient
}

// Do sends a http request with the config by default client
func Do(ctx context.Context, config *Config) (*Response, error) {
	return defaultClient.Do(ctx, config)
}

// Get sends a http get request by default client
func Get(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	return defaultClient.Get(ctx, url, opts...)
}

// Delete sends a http delete request by default client
func Delete(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	return defaultClient.Delete(ctx, url, opts...)
}

// Head sends a http head request by default client
func Head(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	return defaultClient.Head(ctx, url, opts...)
}

// Options sends a http options request by default client
func Options(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	return defaultClient.Options(ctx, url, opts...)
}

// Post sends a http post request by default client
func Post(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return defaultClient.Post(ctx, url, body, opts...)
}

// Put sends a http put request by default client
func Put(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return defaultClient.Put(ctx, url, body, opts...)
}

// Patch sends a http patch request by default client
func Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*Response, error) {
	return defaultClient.Patch(ctx, url, body, opts...)
}

// Upload sends a http post request with multipart form by default client
func Upload(ctx context.Context, url string, form *MultipartForm, opts ...RequestOption) (*Response, error) {
	return defaultClient.Upload(ctx, url, form, opts...)
}

func toJSON[T any](resp *Response, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	result := new(T)
	err = resp.JSON(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func orDefault(c *Client) *Client {
	if c == nil {
		return defaultClient
	}
	return c
}

// DoJSON sends a http request with the config and unmarshals
// the response to T. If the client is nil, the default client is used.
func DoJSON[T any](ctx context.Context, c *Client, config *Config) (*T, error) {
	return toJSON[T](orDefault(c).Do(ctx, config))
}

// GetJSON sends a http get request and unmarshals the response to T.
// If the client is nil, the default client is used.
func GetJSON[T any](ctx context.Context, c *Client, url string, opts ...RequestOption) (*T, error) {
	return toJSON[T](orDefault(c).Get(ctx, url, opts...))
}

// DeleteJSON sends a http delete request and unmarshals the response to T.
// If the client is nil, the default client is used.
func DeleteJSON[T any](ctx context.Context, c *Client, url string, opts ...RequestOption) (*T, error) {
	return toJSON[T](orDefault(c).Delete(ctx, url, opts...))
}

// PostJSON sends a http post request and unmarshals the response to T.
// If the client is nil, the default client is used.
func PostJSON[T any](ctx context.Context, c *Client, url string, body any, opts ...RequestOption) (*T, error) {
	return toJSON[T](orDefault(c).Post(ctx, url, body, opts...))
}

// PutJSON sends a http put request and unmarshals the response to T.
// If the client is nil, the default client is used.
func PutJSON[T any](ctx context.Context, c *Client, url string, body any, opts ...RequestOption) (*T, error) {
	return toJSON[T](orDefault(c).Put(ctx, url, body, opts...))
}

// PatchJSON sends a http patch request and unmarshals the response to T.
// If the client is nil, the default client is used.
func PatchJSON[T any](ctx context.Context, c *Client, url string, body any, opts ...RequestOption) (*T, error) {
	return toJSON[T](orDefault(c).Patch(ctx, url, body, opts...))
}

// UploadJSON sends a http post request with multipart form and unmarshals
// the response to T. If the client is nil, the default client is used.
func UploadJSON[T any](ctx context.Context, c *Client, url string, form *MultipartForm, opts ...RequestOption) (*T, error) {
	return toJSON[T](orDefault(c).Upload(ctx, url, form, opts...))
}

// GetInternalErrorCategory get the category of net op error
func GetInternalErrorCategory(err error) string {
	if errors.Is(err, context.Canceled) {
		return ErrCategoryCanceled
	}
	netErr, ok := err.(net.Error)
	if ok && netErr.Timeout() {
		return ErrCategoryTimeout
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return ErrCategoryDNS
	}
	var addrErr *net.AddrError
	if errors.As(err, &addrErr) {
		return ErrCategoryAddr
	}
	var opErr *net.OpError
	urlErr, ok := netErr.(*url.Error)
	if ok {
		opErr, _ = urlErr.Err.(*net.OpError)
	}

	if opErr == nil {
		opErr, ok = netErr.(*net.OpError)
		if !ok {
			return ""
		}
	}

	switch e := opErr.Err.(type) {
	// 针对以下几种系统调用返回对应类型
	case *os.SyscallError:
		if no, ok := e.Err.(syscall.Errno); ok {
			switch no {
			case syscall.ECONNREFUSED:
				return ErrCategoryRefused
			case syscall.ECONNABORTED:
				return ErrCategoryAborted
			case syscall.ECONNRESET:
				return ErrCategoryReset
			case syscall.ETIMEDOUT:
				return ErrCategoryTimeout
			}
		}
	}

	return ""
}

type Stats struct {
	Route               string `json:"route,omitempty"`
	Method              string `json:"method,omitempty"`
	Result              int    `json:"result,omitempty"`
	URI                 string `json:"uri,omitempty"`
	Status              int    `json:"status,omitempty"`
	Reused              bool   `json:"reused,omitempty"`
	DNSCoalesced        bool   `json:"dnsCoalesced"`
	Addr                string `json:"addr,omitempty"`
	Use                 int    `json:"use,omitempty"`
	DNSUse              int    `json:"dnsUse,omitempty"`
	TCPUse              int    `json:"tcpUse,omitempty"`
	TLSUse              int    `json:"tlsUse,omitempty"`
	RequestSendUse      int    `json:"requestSendUse"`
	ServerProcessingUse int    `json:"serverProcessingUse,omitempty"`
	ContentTransferUse  int    `json:"contentTransferUse,omitempty"`
	Size                int    `json:"size,omitempty"`
}

func ceilToMs(d time.Duration) int {
	if d == 0 {
		return 0
	}
	offset := 0
	if d%time.Millisecond != 0 {
		offset++
	}
	return int(d.Milliseconds()) + offset
}

// GetStats get stats of request
func GetStats(conf *Config, err error) (stats Stats) {
	status := -1
	resp := conf.Response
	size := -1
	if resp != nil {
		status = resp.Status
		size = len(resp.Data)
	}
	result := ResultSuccess
	if err != nil {
		result = ResultFail
	}
	stats = Stats{
		Route:  conf.Route,
		Method: conf.Method,
		Result: result,
		URI:    conf.GetURL(),
		Status: status,
		Size:   size,
	}
	ht := conf.HTTPTrace
	if ht != nil {
		stats.Reused = ht.Reused
		stats.Addr = ht.Addr
		stats.DNSCoalesced = ht.DNSCoalesced
		timelineStats := ht.Stats()
		stats.Use = ceilToMs(timelineStats.Total)

		stats.DNSUse = ceilToMs(timelineStats.DNSLookup)
		stats.TCPUse = ceilToMs(timelineStats.TCPConnection)
		stats.TLSUse = ceilToMs(timelineStats.TLSHandshake)
		stats.RequestSendUse = ceilToMs(timelineStats.RequestSend)
		stats.ServerProcessingUse = ceilToMs(timelineStats.ServerProcessing)
		stats.ContentTransferUse = ceilToMs(timelineStats.ContentTransfer)
	}
	return stats
}

// MapToValues converts map[string]string to url.Values
func MapToValues(data map[string]string) url.Values {
	query := make(url.Values)
	for k, v := range data {
		query.Add(k, v)
	}
	return query
}

// MapToValuesOmitEmpty converts map[string]string to url.Values,
// and omit empty value
func MapToValuesOmitEmpty(data map[string]string) url.Values {
	query := make(url.Values)
	for k, v := range data {
		if v == "" {
			continue
		}
		query.Add(k, v)
	}
	return query
}
