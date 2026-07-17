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
- `HTTPClient` HTTP请求的Client，如果未指定则使用默认值：`http.DefaultClient`，对应`WithHTTPClient`
- `Adapter` 能自定义HTTP请求的处理函数，主要方便各类mock测试场景，对应`WithAdapter`
- `MaxConcurrency` 最大并发请求数，小于0时所有请求均失败，对应`WithMaxConcurrency`，运行时可通过`SetMaxConcurrency`调整
- `EnableTrace` 是否启用事件跟踪，包括HTTP请求中的DNS解析、HTTP发送、开始接收数据等事件，对应`WithEnableTrace`
- `RequestInterceptors` 请求的相关拦截器，对应`WithRequestInterceptor`
- `ResponseInterceptors` 响应的相关拦截器，对应`WithResponseInterceptor`
- `OnError` 当请求出错时回调（列表），可在此处重新对出错封装为自定义出错类型或出错率监控，对应`WithOnError`
- `OnDone` 请求完成时回调（列表），包括成功或失败的请求，用于HTTP请求的相关性能与出错统计，对应`WithOnDone`
- `OnBeforeRequest` 创建新请求前回调（列表），用于在请求前添加一些公共参数，或直接设置`config.Response`短路请求，对应`WithOnBeforeRequest`
- `JSONMarshal`/`JSONUnmarshal` 自定义json编解码函数，对应`WithJSONMarshal`/`WithJSONUnmarshal`

# Config

请求相关的配置参数，其大部分参数与`ClientConfig`一致，在请求发送时，会将`ClientConfig`的参数与`Config`中的合并，优先使用`Config`的参数；client级的拦截器与监听函数在请求级的之前执行。

- `Route` route参数，如：/users/:type
- `URL` 请求的URL
- `Method` HTTP请求类型，默认为`GET`
- `BaseURL` 请求的基本URL，一般在client配置中设置
- `TransformRequest` 请求数据的转换处理，一般在client配置中设置，如果某个请求需要单独处理则再设置此处理函数
- `TransformResponse` 响应数据的转换处理，一般在client配置中设置，如果某个请求需要单独处理则再设置此处理函数
- `Headers` 添加请求头，如果client配置中也有设置，则合并请求头的配置
- `Params` 路由中的参数，此参数用于替换url中的`:key`参数（值会做URL转义）
- `Query` 请求的query参数
- `Body` 请求的实体数据，任意HTTP方法均支持
- `Concurrency` 请求发送时client的并发请求数，此属性自动赋值，不需要设置
- `Timeout` 请求响应超时设置
- `HTTPClient` HTTP请求的Client，如果未指定则使用默认值：`http.DefaultClient`
- `Adapter` 能自定义HTTP请求的处理函数，主要方便各类mock测试场景
- `RequestInterceptors` 请求的相关拦截器
- `ResponseInterceptors` 响应的相关拦截器
- `OnError` 当请求出错时回调（列表）
- `OnDone` 请求完成时回调（列表）
- `OnBeforeRequest` 创建新请求前回调（列表）
- `HTTPTrace` 启用trace之后生成的http调用链中的相关记录

注意：v2中context不再是Config的字段，而是所有请求方法的首个参数，可通过`config.Context()`读取。
