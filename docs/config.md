---
description: axios的相关配置参数
---

# ClientConfig

client的公共参数配置，通过`NewClient(opts ...ClientOption)`的函数式选项或`NewClientWithConfig(config ClientConfig)`设置，用于指定服务的各类公共处理与处理函数，如超时、请求头以及请求数据转换、响应数据转换等。client构造后配置不可变，并发安全。

- `BaseURL` client中所有请求的基本URL，最终请求的地址将是BaseURL + URL，对应`WithBaseURL`
- `TransformRequest` 请求数据的转换处理，当Body不为nil时将其转换为字节，默认的transform根据数据类型转换为`x-www-form-urlencoded`或者`json`，一般使用中只需要使用默认处理则可，对应`WithTransformRequest`
- `TransformResponse` 响应数据的转换处理，默认的响应转换支持解压`gzip`以及`br`，对应`WithTransformResponse`
- `Headers` 添加公共的请求头，对应`WithHeader`/`WithHeaders`
- `Timeout` 请求响应超时设置，对应`WithTimeout`
- `MaxBodySize` 响应体最大字节数，`0` 表示不限制，对应`WithMaxBodySize`；超出时返回 `*Error`（`errors.Is(err, axios.ErrBodyTooLarge)`）
- `ValidateStatus` 响应状态码校验函数，为 `nil` 时接受所有状态码（默认），对应`WithValidateStatus`；可使用内置 `ValidateStatus2xx` 仅接受 2xx。校验失败时仍返回 `*Response`，同时 `error` 为 `*Error`（`errors.Is(err, axios.ErrStatusInvalid)`）
- `Retry` 默认重试策略，对应`WithRetry`；请求级可用 `RetryPolicy` 覆盖。详见下方 Retry
- `HTTPClient` HTTP请求的Client；未指定时为每个 Client 创建**私有** `http.Client`（独立连接池，Transport 从 `http.DefaultTransport` clone），对应`WithHTTPClient`
- `Adapter` 能自定义HTTP请求的处理函数，主要方便各类mock测试场景，对应`WithAdapter`
- `MaxConcurrency` 最大并发请求数，小于0时所有请求均失败，对应`WithMaxConcurrency`，运行时可通过`SetMaxConcurrency`调整
- `EnableTrace` 是否启用事件跟踪，包括HTTP请求中的DNS解析、HTTP发送、开始接收数据等事件，对应`WithEnableTrace`
- `RequestInterceptors` 请求的相关拦截器，对应`WithRequestInterceptor`
- `ResponseInterceptors` 响应的相关拦截器，对应`WithResponseInterceptor`
- `OnError` 当请求出错时回调（列表），可在此处重新对出错封装为自定义出错类型或出错率监控，对应`WithOnError`
- `OnDone` 请求完成时回调（列表），包括成功或失败的请求，用于HTTP请求的相关性能与出错统计，对应`WithOnDone`
- `OnBeforeRequest` 创建新请求前回调（列表），用于在请求前添加一些公共参数，或直接设置`config.Response`短路请求，对应`WithOnBeforeRequest`
- `JSONMarshal`/`JSONUnmarshal` 自定义json编解码函数，对应`WithJSONMarshal`/`WithJSONUnmarshal`

`GetConfig()` 返回配置的**深拷贝**（Headers、Retry 与各类 interceptor/listener 切片），修改返回值不会影响 Client；`HTTPClient` / `Adapter` / 函数值仍按引用共享。

## Clone

`client.Clone(opts ...ClientOption) *Client` 基于现有 Client 深拷贝配置并叠加选项，得到新 Client：

- Headers / Retry / interceptors 等为拷贝，互不影响
- **默认共享**源 Client 的 `HTTPClient`（共用连接池）；需要隔离时再 `WithHTTPClient(...)`
- Mock 状态与进行中的并发计数**不会**拷贝

```go
api := axios.NewClient(axios.WithBaseURL("https://api.example.com"), axios.WithTimeout(10*time.Second))
users := api.Clone(axios.WithBaseURL("https://api.example.com/users"))
```

## Retry

`RetryConfig` 字段：

| 字段 | 说明 |
|------|------|
| `MaxRetries` | 首次之后的重试次数；总尝试次数 = MaxRetries + 1；`0` 表示不重试 |
| `MinWait` | 首次重试前基础等待，默认 100ms |
| `MaxWait` | 退避上限，`0` 表示不限制 |
| `Multiplier` | 指数退避倍数，默认 2 |
| `Methods` | 允许重试的方法；空则为 GET / HEAD / OPTIONS |
| `Decide` | 自定义是否重试；`nil` 时用 `DefaultRetryIf` |

`DefaultRetryIf` 会重试：超时 / DNS / 连接 refused·reset·aborted，以及状态码 408、429、500、502、503、504（需配合 `ValidateStatus` 才会产生状态错误）。**不会**重试 `context.Canceled`、`ErrBodyTooLarge`、其它 4xx。

**Body 可重放**：`io.Reader` 类型 body 只会被读一次，此类请求**不会**重试。请使用 `[]byte`、`string`、`url.Values` 或会每次重新 marshal 的结构体/map。

`Config.Attempt` 为最后一次尝试的 1-based 序号，可在 `OnDone` 中观测。

```go
c := axios.NewClient(
    axios.WithValidateStatus(axios.ValidateStatus2xx),
    axios.WithRetry(axios.DefaultRetryConfig()),
)
// 或单次：c.Get(ctx, "/x", axios.RetryPolicy(axios.RetryConfig{MaxRetries: 1, MinWait: time.Millisecond}))
```

# Config

请求相关的配置参数，其大部分参数与`ClientConfig`一致，在请求发送时，会将`ClientConfig`的参数与`Config`中的合并，优先使用`Config`的参数；client级的拦截器与监听函数在请求级的之前执行。

- `Route` route参数，如：/users/:type
- `URL` 请求的URL
- `Method` HTTP请求类型，默认为`GET`
- `BaseURL` 请求的基本URL，一般在client配置中设置
- `TransformRequest` 请求数据的转换处理，一般在client配置中设置，如果某个请求需要单独处理则再设置此处理函数
- `TransformResponse` 响应数据的转换处理，一般在client配置中设置，如果某个请求需要单独处理则再设置此处理函数
- `Headers` 添加请求头，如果client配置中也有设置，则合并请求头的配置
- `Params` 路由参数：仅替换 **path 中整段** 形如 `:key` 的 segment（值会做 PathEscape）；不会误伤 query/fragment，也不会出现 `:id` 覆盖 `:idExtra` 的前缀问题
- `Query` 请求的query参数
- `Body` 请求的实体数据，任意HTTP方法均支持
- `Concurrency` 请求发送时client的并发请求数，此属性自动赋值，不需要设置
- `Timeout` 请求响应超时设置
- `MaxBodySize` 响应体最大字节数，`0` 表示不限制；请求级可通过 `MaxBodySize(n)` 设置
- `ValidateStatus` 状态码校验；请求级可通过 `StatusValidator(fn)` 设置
- `Retry` 重试策略；请求级可通过 `RetryPolicy(cfg)` 设置
- `Attempt` 最后一次尝试序号（1-based），由客户端写入
- `HTTPClient` HTTP请求的Client
- `Adapter` 能自定义HTTP请求的处理函数，主要方便各类mock测试场景
- `RequestInterceptors` 请求的相关拦截器
- `ResponseInterceptors` 响应的相关拦截器
- `OnError` 当请求出错时回调（列表）
- `OnDone` 请求完成时回调（列表）
- `OnBeforeRequest` 创建新请求前回调（列表）
- `HTTPTrace` 启用trace之后生成的http调用链中的相关记录

注意：v2中context不再是Config的字段，而是所有请求方法的首个参数，可通过`config.Context()`读取。

默认 adapter 会在读取响应时流式解压 `gzip` / `br`（只缓冲一份解码后的数据）；`DefaultTransformResponse` 仍保留作为自定义 adapter 的兜底解压。
