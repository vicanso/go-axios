---
description: request的各种使用方式
---

## Request(config *Config) (resp *Response, err error)

通过Config的各参数配置(具体各属性参考Config的说明)，指定HTTP请求。

```go
import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/vicanso/go-axios"
)

func main() {
	resp, err := axios.Request(&axios.Config{
		URL: "https://tiny.npmtrend.com/users/v1/me/login",
	})
	fmt.Println(err)
	fmt.Println(resp)
}
```

## Get(url string, query ...url.Values) (resp *Response, err error)

HTTP GET请求，query部分为可选参数，需要注意需要query为不定长但只支持不传或者只传一个参数。

## Delete(url string, query ...url.Values) (resp *Response, err error)

HTTP DELETE请求，query部分为可选参数，需要注意需要query为不定长但只支持不传或者只传一个参数。

## Head(url string, query ...url.Values) (resp *Response, err error) 

HTTP HEAD请求，query部分为可选参数，需要注意需要query为不定长但只支持不传或者只传一个参数。

## Options(url string, query ...url.Values) (resp *Response, err error)

HTTP OPTIONS请求，query部分为可选参数，需要注意需要query为不定长但只支持不传或者只传一个参数。

## Post(url string, data interface{}, query ...url.Values) (resp *Response, err error) 

HTTP POST请求，默认的TransformRequest中，对于data如果是`[]byte`或者`string`则转换为`[]byte`直接请求。如果是`url.Values`，调用`Encode`方法之后设置为`application/x-www-form-urlencoded;charset=utf-8`。其它类型则使用`jsonMarshal`转换为字节，并设置`application/json;charset=utf-8`。query部分为可选参数，需要注意需要query为不定长但只支持不传或者只传一个参数。

## Patch(url string, data interface{}, query ...url.Values) (resp *Response, err error)

HTTP PATCH请求，处理方式与`POST`一致。

## Put(url string, data interface{}, query ...url.Values) (resp *Response, err error)

HTTP PUT请求，处理方式与`POST`一致。