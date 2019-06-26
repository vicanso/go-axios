# go-axios

HTTP client for golang, it derives [axios](https://github.com/axios/axios).

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
- `EnableTrace` enable trace
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

## examples

```go
package main

import (
	"fmt"
	"time"

	"github.com/vicanso/go-axios"
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
	ins := axios.NewInstance(&axios.InstanceConfig{
		BaseURL:     "https://www.baidu.com/",
		Timeout:     time.Second,
		EnableTrace: true,
		ResponseInterceptors: []axios.ResponseInterceptor{
			httpStats,
		},
	})
	_, err := ins.Get("/")
	if err != nil {
		panic(err)
	}
}
```