# go-axios

[![Build Status](https://img.shields.io/travis/vicanso/go-axios.svg?label=linux+build)](https://travis-ci.org/vicanso/go-axios)

HTTP client for golang, it derives [axios](https://github.com/axios/axios).

```go
package main

import (
	"fmt"

	"github.com/vicanso/go-axios"
)

func main() {
	ins := axios.NewInstance(nil)
	resp, err := ins.Get("https://aslant.site/")
	fmt.Println(resp.Status, err)
}
```

## API

### Config

The http request config.

- `URL` http request url.
- `Method` http request method, default is `GET`.
- `BaseURL` will be prepended to `url` unless `url` is absolute.
- `TransformRequest` convert request body, it's only for request methods `POST`, `PUT` and `PATCH`, default is `DefaultTransformRequest`.
- `TransformResponse` convert response body, default is `DefaultTransformResponse`, it support `gzip` and `br` encoding.
- `Headers` custom headers for request.
- `Params` will be replaced to `url`.
- `Query` will be appended to `url` as `querystring`.
- `Body` will be sent to server, it's only for request methods `POST`, `PUT` and `PATCH`.
- `Timeout` if the request takes longer than `timeout`, the request will be aborted.
- `Client` http client, default is `http.DefaultCleint`.
- `Adapter` allows custom handling of requests which makes testing easier.
- `RequestInterceptors` request interceptor list.
- `ResponseInterceptors` response interceptor list.
- `OnError` on error event.

### InstanceConfig

The http instance config, it will be merged to http request config.

- `BaseURL` will be prepended to `url` unless `url` is absolute.
- `TransformRequest` convert request body, it's only for request methods `POST`, `PUT` and `PATCH`, default is `DefaultTransformRequest`.
- `TransformResponse` convert response body, default is `DefaultTransformResponse`, it support `gzip` and `br` encoding.
- `Headers` custom headers for request.`
- `Timeout` if the request takes longer than `timeout`, the request will be aborted.
- `Client` http client, default is `http.DefaultCleint`.
- `Adapter` allows custom handling of requests which makes testing easier.
- `RequestInterceptors` request interceptor list.
- `ResponseInterceptors` response interceptor list.
- `EnableTrace` enable trace.
- `OnError` on error event.

### Request(config *Config) (resp *Response, err error)

HTTP requset.

```go
package main

import (
	"fmt"

	"github.com/vicanso/go-axios"
)

func main() {
	ins := axios.NewInstance(nil)
	resp, err := ins.Request(&axios.Config{
		URL: "https://aslant.site/",
	})
	fmt.Println(err)
	fmt.Println(resp)
}
```

### Get(url string) (resp *Response, err error)

HTTP get request.


### Delete(url string) (resp *Response, err error)

HTTP delete request.

### Head(url string) (resp *Response, err error) 

HTTP head request.

### Options(url string) (resp *Response, err error)

HTTP options request.

### Post(url string, data interface{}) (resp *Response, err error) 

HTTP post request.

### Put(url string, data interface{}) (resp *Response, err error)

HTTP put request.

### Patch(url string, data interface{}) (resp *Response, err error)

HTTP patch request.

## Examples

### Request Stats
Add stats for all request

```go
package main

import (
	"fmt"

	"github.com/vicanso/go-axios"
)

var (
	aslant = axios.NewInstance(&axios.InstanceConfig{
		BaseURL:     "https://aslant.site/",
		EnableTrace: true,
		ResponseInterceptors: []axios.ResponseInterceptor{
			httpStats,
		},
	})
)

func httpStats(resp *axios.Response) (err error) {
	stats := make(map[string]interface{})
	config := resp.Config
	stats["url"] = config.URL
	stats["status"] = resp.Status

	ht := config.HTTPTrace
	if ht != nil {
		stats["timeline"] = config.HTTPTrace.Stats()
		stats["addr"] = ht.Addr
		stats["reused"] = ht.Reused
	}
	// 可以将相应的记录写入统计数据
	fmt.Println(stats)
	return nil
}

func main() {
	resp, err := aslant.Get("/")
	fmt.Println(err)
	fmt.Println(resp.Status)
}

```

### Convert Error

Convert response(4xx, 5xx) to an error.

```go
package main

import (
	"errors"
	"fmt"

	"github.com/vicanso/go-axios"

	jsoniter "github.com/json-iterator/go"
)

var (
	standardJSON = jsoniter.ConfigCompatibleWithStandardLibrary
)
var (
	aslant = axios.NewInstance(&axios.InstanceConfig{
		BaseURL: "https://ip.aslant.site/",
		ResponseInterceptors: []axios.ResponseInterceptor{
			convertResponseToError,
		},
	})
)

// convertResponseToError convert http response(4xx, 5xx) to error
func convertResponseToError(resp *axios.Response) (err error) {
	if resp.Status >= 400 {
		message := standardJSON.Get(resp.Data, "message").ToString()
		if message == "" {
			message = "Unknown Error"
		}
		// or you can use custom error
		err = errors.New(message)
	}
	return
}

func main() {
	_, err := aslant.Get("/ip-locations/json/123")
	fmt.Println(err)
}
```

### Mock For Test

Mock request for test.

```go
package main

import (
	"fmt"

	"github.com/vicanso/go-axios"
)

type (
	// UserInfo user info
	UserInfo struct {
		Account string `json:"account,omitempty"`
		Name    string `json:"name,omitempty"`
	}
)

var (
	aslant = axios.NewInstance(&axios.InstanceConfig{
		BaseURL: "https://aslant.site/",
	})
)

// getUserInfo get user info from aslant.site
func getUserInfo() (userInfo *UserInfo, err error) {
	resp, err := aslant.Get("/users/me")
	if err != nil {
		return
	}
	userInfo = new(UserInfo)
	err = resp.JSON(userInfo)
	if err != nil {
		return
	}
	return
}

// mockUserInfo mock user info
func mockUserInfo(data []byte) (done func()) {
	originalAdapter := aslant.Config.Adapter
	aslant.Config.Adapter = func(config *axios.Config) (resp *axios.Response, err error) {
		resp = &axios.Response{
			Data:   data,
			Status: 200,
		}
		return
	}

	done = func() {
		aslant.Config.Adapter = originalAdapter
	}
	return
}

func main() {
	mockUserInfo([]byte(`{"account":"tree", "name":"tree.xie"}`))
	userInfo, err := getUserInfo()
	fmt.Println(err)
	fmt.Println(userInfo)
}
```