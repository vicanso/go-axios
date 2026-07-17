---
description: request的各种使用方式
---

# Request

v2中所有请求方法均以`context.Context`作为首个参数，请求级参数通过`RequestOption`传入。

## Do(ctx context.Context, config *Config) (resp *Response, err error)

通过Config的各参数配置(具体各属性参考Config的说明)，指定HTTP请求。

```go
import (
	"context"
	"fmt"

	axios "github.com/vicanso/go-axios/v2"
)

func main() {
	resp, err := axios.Do(context.Background(), &axios.Config{
		URL: "https://tiny.npmtrend.com/users/v1/me/login",
	})
	fmt.Println(err)
	fmt.Println(resp)
}
```

## Get / Delete / Head / Options

`Get(ctx context.Context, url string, opts ...RequestOption) (resp *Response, err error)`

HTTP GET请求，query、路由参数、请求头等通过`RequestOption`指定。`Delete`、`Head`与`Options`的使用方式一致。

```go
resp, err := ins.Get(ctx, "/users/:type",
	axios.Param("type", "vip"),
	axios.Query("cache", "false"),
	axios.Header("X-Request-Id", "1"),
	// 可选：单次重试策略（覆盖 client 默认）
	// axios.RetryPolicy(axios.DefaultRetryConfig()),
)
```

## Post / Put / Patch

`Post(ctx context.Context, url string, body any, opts ...RequestOption) (resp *Response, err error)`

HTTP POST请求，默认的TransformRequest中，对于body如果是`[]byte`或者`string`则转换为`[]byte`直接请求。如果是`url.Values`，调用`Encode`方法之后设置为`application/x-www-form-urlencoded;charset=utf-8`。其它类型则使用json marshal转换为字节，并设置`application/json;charset=utf-8`。`Put`与`Patch`的处理方式一致。

## Upload

`Upload(ctx context.Context, url string, form *MultipartForm, opts ...RequestOption) (resp *Response, err error)`

以multipart form的形式上传文件：

```go
form := axios.NewMultipartForm()
_ = form.AddFile("file", "test.txt", data)
_ = form.AddField("type", "vip")
resp, err := ins.Upload(ctx, "/upload", form)
```

## 泛型JSON方法

v1中的`EnhanceXXX`系列方法被泛型函数替代，直接返回unmarshal后的结构体指针。client传`nil`时使用默认client：

- `GetJSON[T any](ctx context.Context, c *Client, url string, opts ...RequestOption) (*T, error)`
- `DeleteJSON[T any](ctx context.Context, c *Client, url string, opts ...RequestOption) (*T, error)`
- `PostJSON[T any](ctx context.Context, c *Client, url string, body any, opts ...RequestOption) (*T, error)`
- `PutJSON[T any]` / `PatchJSON[T any]` 与`PostJSON`一致
- `UploadJSON[T any](ctx context.Context, c *Client, url string, form *MultipartForm, opts ...RequestOption) (*T, error)`
- `DoJSON[T any](ctx context.Context, c *Client, config *Config) (*T, error)`

```go
type UserInfo struct {
	Account string `json:"account"`
}

user, err := axios.GetJSON[UserInfo](ctx, ins, "/users/me")
if err != nil {
	return err
}
fmt.Println(user.Account)
```

## RequestOption

- `Query(key, value string)` 添加query参数
- `QueryValues(query url.Values)` 添加url.Values中所有query参数
- `QueryMap(query map[string]string)` 添加map中所有query参数
- `Param(key, value string)` 设置路由参数，替换url中的`:key`（值会做URL转义）
- `Header(key, value string)` 添加请求头
- `Timeout(d time.Duration)` 设置单个请求的超时
- `Route(route string)` 设置route（用于统计与MultiMock）
- `Body(body any)` 设置请求体（任意方法均支持）
- `EnableTrace()` 启用该请求的http trace
