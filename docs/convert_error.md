---
description: 接口请求出错的处理
---

# Error

默认的出错只针对HTTP请求失败，对于状态码为4xx，5xx的请求，需要针对各服务从响应数据中获取对应的出错信息，再转换为相应的自定义出错对象。在IPLocation这个服务，出错时响应的数据为`{"category": "出错类别", "message": "出错信息"}，因此增加自定义出错转换，代码如下：


```go
package main

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/vicanso/go-axios"
	"github.com/vicanso/hes"
)

var (
	ipLocationIns *axios.Instance
)

// convertResponseToError convert http response(4xx, 5xx) to error
func convertResponseToError(resp *axios.Response) error {
	if resp.Status >= 400 {
		he := &hes.Error{}
		err := resp.JSON(he)
		if err != nil {
			he.Message = string(resp.Data)
		}
		return he
	}
	return nil
}

func main() {
	_, err := ipLocationIns.Get("/ip-locations/json/123")
	fmt.Println(err)
}

func init() {
	headers := make(http.Header)
	headers.Set("Accept", "application/json")
	ipLocationIns = axios.NewInstance(&axios.InstanceConfig{
		Headers: headers,
		BaseURL: "https://ip.npmtrend.com",
		ResponseInterceptors: []axios.ResponseInterceptor{
			convertResponseToError,
		},
	})
}
```