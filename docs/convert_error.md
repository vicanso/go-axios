---
description: 接口请求出错的处理
---

# Error

默认情况下，HTTP 传输成功即 `err == nil`，**不会**因 4xx/5xx 自动失败。

## 结构化错误 `*axios.Error`

库在以下场景返回 `*axios.Error`（可用 `errors.As` / `errors.Is`）：

| 场景 | `errors.Is` | 说明 |
|------|-------------|------|
| 状态码校验失败 | `axios.ErrStatusInvalid` | 配置了 `ValidateStatus` 且返回 false；`Response` 仍可用 |
| 响应体超限 | `axios.ErrBodyTooLarge` | 超过 `MaxBodySize`；`Response.Data` 为截断后的数据 |

```go
resp, err := c.Get(ctx, "/users/me")
if err != nil {
	var axErr *axios.Error
	if errors.As(err, &axErr) {
		fmt.Println(axErr.Status, axErr.Response)
	}
}
```

## 仅接受 2xx

```go
c := axios.NewClient(
	axios.WithBaseURL("https://api.example.com"),
	axios.WithValidateStatus(axios.ValidateStatus2xx),
)
// 或单次请求：
// c.Get(ctx, "/x", axios.StatusValidator(axios.ValidateStatus2xx))
```

## 按业务体转换 4xx/5xx

若需要从响应 JSON 解析业务错误（而非仅状态码），仍推荐 ResponseInterceptor：

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
