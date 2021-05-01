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
	"os"
	"syscall"
	"testing"
	"time"

	HT "github.com/vicanso/http-trace"

	"github.com/stretchr/testify/assert"
)

func TestSetJSONMarshal(t *testing.T) {
	marshal := func(v interface{}) ([]byte, error) {
		return json.Marshal(v)
	}
	unmarshal := func(data []byte, v interface{}) error {
		return json.Unmarshal(data, v)
	}
	SetJSONMarshal(marshal)
	SetJSONUnmarshal(unmarshal)
}

func TestDefaultInstance(t *testing.T) {
	assert := assert.New(t)
	ins := GetDefaultInstance()
	original := ins.Config.Adapter
	defer func() {
		ins.Config.Adapter = original
	}()
	ins.Config.Adapter = func(conf *Config) (*Response, error) {
		return &Response{
			Config: conf,
		}, nil
	}
	resp, err := Request(&Config{})
	assert.Nil(err)
	assert.Equal("GET", resp.Config.Method)

	resp, err = Get("/")
	assert.Nil(err)
	assert.Equal("GET", resp.Config.Method)

	resp, err = Delete("/")
	assert.Nil(err)
	assert.Equal("DELETE", resp.Config.Method)

	resp, err = Head("/")
	assert.Nil(err)
	assert.Equal("HEAD", resp.Config.Method)

	resp, err = Options("/")
	assert.Nil(err)
	assert.Equal("OPTIONS", resp.Config.Method)

	resp, err = Post("/", nil)
	assert.Nil(err)
	assert.Equal("POST", resp.Config.Method)

	resp, err = Put("/", nil)
	assert.Nil(err)
	assert.Equal("PUT", resp.Config.Method)

	resp, err = Patch("/", nil)
	assert.Nil(err)
	assert.Equal("PATCH", resp.Config.Method)
}

type timeoutErr struct {
	error
}

func (te *timeoutErr) Timeout() bool {
	return true
}

func (te *timeoutErr) Temporary() bool {
	return false
}

func TestGetErrorCategory(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(ErrCategoryCanceled, GetInternalErrorCategory(context.Canceled))

	assert.Equal(ErrCategoryTimeout, GetInternalErrorCategory(&timeoutErr{}))

	assert.Equal(ErrCategoryDNS, GetInternalErrorCategory(&net.DNSError{}))

	assert.Equal(ErrCategoryAddr, GetInternalErrorCategory(&net.AddrError{}))

	assert.Equal(ErrCategoryRefused, GetInternalErrorCategory(&net.OpError{
		Err: &os.SyscallError{
			Err: syscall.ECONNREFUSED,
		},
	}))
	assert.Equal(ErrCategoryAborted, GetInternalErrorCategory(&net.OpError{
		Err: &os.SyscallError{
			Err: syscall.ECONNABORTED,
		},
	}))
	assert.Equal(ErrCategoryReset, GetInternalErrorCategory(&net.OpError{
		Err: &os.SyscallError{
			Err: syscall.ECONNRESET,
		},
	}))
	assert.Equal(ErrCategoryTimeout, GetInternalErrorCategory(&net.OpError{
		Err: &os.SyscallError{
			Err: syscall.ETIMEDOUT,
		},
	}))

	assert.Empty(GetInternalErrorCategory(errors.New("a")))
}

func TestGetStats(t *testing.T) {
	ht := &HT.HTTPTrace{
		Addr:    "1.1.1.1:80",
		Reused:  true,
		Start:   time.Unix(1, 0),
		GetConn: time.Unix(2, 0),

		DNSStart: time.Unix(1, 0),
		DNSDone:  time.Unix(3, 0),

		ConnectStart: time.Unix(1, 0),
		ConnectDone:  time.Unix(4, 0),

		TLSHandshakeStart: time.Unix(1, 0),
		TLSHandshakeDone:  time.Unix(5, 0),

		GotConnect:           time.Unix(1, 0),
		GotFirstResponseByte: time.Unix(6, 0),

		Done: time.Unix(12, 0),
	}
	assert := assert.New(t)
	conf := &Config{
		Method:  "GET",
		BaseURL: "http://127.0.0.1",
		Route:   "/users/v1/:type",
		URL:     "/users/v1/me",
		Response: &Response{
			Status: 400,
			Data:   []byte("test"),
		},
		HTTPTrace: ht,
	}

	stats := GetStats(conf, errors.New("fail"))

	assert.Equal("/users/v1/:type", stats.Route)
	assert.Equal("GET", stats.Method)
	assert.Equal(ResultFail, stats.Result)
	assert.Equal("http://127.0.0.1/users/v1/me", stats.URI)
	assert.Equal(400, stats.Status)
	assert.True(stats.Reused)
	assert.Equal("1.1.1.1:80", stats.Addr)
	assert.Equal(2000, stats.DNSUse)
	assert.Equal(3000, stats.TCPUse)
	assert.Equal(4000, stats.TLSUse)
	assert.Equal(5000, stats.ServerProcessingUse)
	assert.Equal(6000, stats.ContentTransferUse)
	assert.Equal(11000, stats.Use)

	assert.Equal(4, stats.Size)
}
