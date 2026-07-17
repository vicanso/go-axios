# go-axios

[![Build Status](https://github.com/vicanso/go-axios/workflows/Test/badge.svg)](https://github.com/vicanso/go-axios/actions)

简单易用的HTTP客户端，参考[axios](https://github.com/axios/axios)的相关实现，支持各类不同的`interceptor`与`transform`，特性如下：

- 基于泛型的响应数据unmarshal，无需手动处理json
- 所有请求方法均以`context.Context`作为首个参数，符合Go惯例
- 通过函数式选项（`ClientOption`/`RequestOption`）配置客户端与请求
- 支持自定义的transform，可根据应用场景指定各类不同的参数格式
- 支持Request与Response的interceptor，可针对请求参数与响应数据添加各类的转换处理
- 默认支持gzip与br两种压缩方式（adapter 流式解压，单次缓冲），节约带宽占用
- 可选 `ValidateStatus`（如 `ValidateStatus2xx`）与 `MaxBodySize` 响应体上限
- 可选重试（`WithRetry` / `DefaultRetryConfig`），支持指数退避与 body 可重放约定
- `Client.Clone` 派生配置（默认共享连接池）
- 结构化错误 `*Error`，便于 `errors.As` / `errors.Is`
- 未指定时为每个 Client 创建私有 `http.Client`（独立连接池）
- 支持启用请求中的各事件记录，可细化请求的各场景耗时
- 简单易用的Mock处理（并发安全）
- Error与Done事件的支持，通过监听事件可快捷收集服务的出错与性能统计

## v2 破坏性变更

v2的module路径为`github.com/vicanso/go-axios/v2`，主要变更：

- `Instance` 更名为 `Client`，通过 `NewClient(opts ...ClientOption)` 创建，构造后配置不可变（并发安全）
- 所有请求方法均需传入 `context.Context`，原 `GetX`/`EnhanceGet` 等方法被移除
- `EnhanceXXX` 系列由泛型函数替代：`GetJSON[T]`、`PostJSON[T]`、`DoJSON[T]` 等
- 请求级参数通过 `RequestOption` 传入：`Query`、`Param`、`Header`、`Timeout`、`Route`、`Body`、`EnableTrace`
- `Config.Client` 更名为 `Config.HTTPClient`；`Config.Context` 移除（由参数传入）
- `TransformRequest`/`TransformResponse` 函数签名增加 `*Config` 参数
- `NewMultipartFile` 更名为 `NewMultipartForm`，返回导出类型 `*MultipartForm`
- 全局的 `SetJSONMarshal`/`SetJSONUnmarshal` 移除，改为 `WithJSONMarshal`/`WithJSONUnmarshal` 客户端选项
- 任意请求方法均支持Body（不再限于POST/PUT/PATCH）
- 修复：默认`Accept`请求头拼写错误、并发数配置的data race、Mock的data race、curl命令的单引号转义、路由参数未做URL转义

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	axios "github.com/vicanso/go-axios/v2"
)

type UserInfo struct {
	Account string `json:"account"`
}

func main() {
	ctx := context.Background()

	// 使用默认的client
	resp, err := axios.Get(ctx, "https://www.baidu.com")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Status)

	// 自定义client，可指定http client、interceptor等
	ins := axios.NewClient(
		axios.WithBaseURL("https://aslant.site"),
		axios.WithEnableTrace(),
		// 可选：限制响应体大小；超出返回 errors.Is(err, axios.ErrBodyTooLarge)
		axios.WithMaxBodySize(8<<20),
		// 可选：非 2xx 视为错误（仍返回 *Response，err 为 *axios.Error）
		axios.WithValidateStatus(axios.ValidateStatus2xx),
		// 可选：幂等请求失败时重试（默认 GET/HEAD/OPTIONS）
		axios.WithRetry(axios.DefaultRetryConfig()),
		axios.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}),
		// 超时设置，建议设置此字段避免接口无超时处理
		axios.WithTimeout(10*time.Second),
		// OnDone 无论成功或失败均会调用，可在此处添加统计
		axios.WithOnDone(func(config *axios.Config, resp *axios.Response, err error) {
			fmt.Println(axios.GetStats(config, err))
		}),
	)

	// 请求级参数：路由参数、query等
	resp, err = ins.Get(ctx, "/users/:type",
		axios.Param("type", "me"),
		axios.Query("cache", "false"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Status)

	// 泛型：直接返回unmarshal后的结构体
	user, err := axios.GetJSON[UserInfo](ctx, ins, "/users/me")
	if err != nil {
		panic(err)
	}
	fmt.Println(user.Account)
}
```
