---
description: 接口请求出错的处理
---

# Error

默认的出错只针对HTTP请求失败，对于状态码为4xx、5xx的请求，需要针对各服务从响应数据中获取对应的出错信息，再转换为相应的自定义出错对象。在IPLocation这个服务，出错时响应的数据为`{"category": "出错类别", "message": "出错信息"}`，因此增加自定义出错转换，代码如下：

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	axios "github.com/vicanso/go-axios/v2"
	"github.com/vicanso/hes"
)

var ipLocationIns *axios.Client

// convertResponseToError convert http response(4xx, 5xx) to error
func convertResponseToError(resp *axios.Response) error {
	if !resp.OK() {
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
	_, err := ipLocationIns.Get(context.Background(), "/ip-locations/json/123")
	fmt.Println(err)
}

func init() {
	headers := make(http.Header)
	headers.Set("Accept", "application/json")
	ipLocationIns = axios.NewClient(
		axios.WithBaseURL("https://ip.npmtrend.com"),
		axios.WithHeaders(headers),
		axios.WithResponseInterceptor(convertResponseToError),
	)
}
```
