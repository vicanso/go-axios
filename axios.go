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
	"fmt"
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
	JSONMarshal func(interface{}) ([]byte, error)
	// JSONUnmarshal json unmarshal function type
	JSONUnmarshal func([]byte, interface{}) error
)

// defaultTimeout is for all request which not set timeout
var (
	defaultIns = NewInstance(&InstanceConfig{
		Timeout: time.Minute,
	})
	// default json marshal
	jsonMarshal = json.Marshal
	// default json unmarshal
	jsonUnmarshal = json.Unmarshal
)

// SetJSONMarshal set json marshal function
func SetJSONMarshal(fn JSONMarshal) {
	jsonMarshal = fn
}

// SetJSONUnmarshal set json unmarshal function
func SetJSONUnmarshal(fn JSONUnmarshal) {
	jsonUnmarshal = fn
}

// Request http request by default instance
func Request(config *Config) (resp *Response, err error) {
	return defaultIns.Request(config)
}

// Get http get request by default instance
func Get(url string, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Get(url, query...)
}

// Delete http delete request by default instance
func Delete(url string, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Delete(url, query...)
}

// Head http head request by default instance
func Head(url string, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Head(url, query...)
}

// Options http options request by default instance
func Options(url string, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Options(url, query...)
}

// Post http post request by default instance
func Post(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Post(url, data, query...)
}

// Put http put request by default instance
func Put(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Put(url, data, query...)
}

// Patch http patch request by default instance
func Patch(url string, data interface{}, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Patch(url, data, query...)
}

// Upload upload file by default instance
func Upload(url string, file *multipartFile, query ...url.Values) (resp *Response, err error) {
	return defaultIns.Upload(url, file, query...)
}

// GetDefaultInstance get default instance
func GetDefaultInstance() *Instance {
	return defaultIns
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
		fmt.Println(timelineStats)
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
