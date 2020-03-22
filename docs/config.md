---
description: axios的相关配置参数
---

# InstanceConfig

实例的公共参数配置，用于指定服务的各类公共处理与处理函数，如超时、请求头以及请求数据转换、响应数据转换等。

- `BaseURL` 实例中所有请求的基本URL，最终请求的地址将是BaseURL + URL
- `TransformRequest` 请求数据的转换处理，针对`POST`，`PATCH`以及`PUT`请求的发送数据转换为字节，默认的transform根据数据类型转换为`x-www-form-urlencoded`或者`json`，一般使用中只需要使用默认处理则可。
- `TransformResponse` 响应数据的转换处理，默认的响应转换支持解压`gzip`以及`br`
- `Headers` 添加公共的请求头
- `Timeout` 请求响应超时设置
- `Client` HTTP请求的Client，如果未指定则使用默认值：`http.DefaultCleint`
- `Adapter` 能自定义HTTP请求的处理函数，主要方便各类mock测试场景
- `RequestInterceptors` 请求的相关拦截器
- `ResponseInterceptors` 响应的相关拦截器
- `EnableTrace` 是否启用事件跟踪，包括HTTP请求中的DNS解析、HTTP发送、开始接收数据等事件
- `OnError` 当请求出错时回调，可在此处重新对出错封装为自定义出错类型或出错率监控
- `OnDone` 请求完成时回调，包括成功或失败的请求，用于HTTP请求的相关性能与出错统计

# Config

请求相关的配置参数，其大部分参数与`InstanceConfig`一致，在请求发送时，会将`InstanceConfig`的参数与`Config`中的合并，优先使用`Config`的参数。

- `URL` 请求的URL
- `Method` HTTP请求类型，默认为`GET`
- `BaseURL` 请求的基本URL，一般在实例配置中设置
- `TransformRequest` 请求数据的转换处理，一般在实例配置中设置，如果某个请求需要单独处理则再设置此处理函数
- `TransformResponse` 响应数据的转换处理，一般在实例配置中设置，如果某个请求需要单独处理则再设置此处理函数
- `Headers` 添加请求头，如果实例配置中也有设置，则合并请求头的配置
- `Params` 路由中的参数，此参数用于替换url中的`:key`参数
- `Query` 请求的query参数
- `Body` 请求的实体数据，用于`POST`，`PUT`以及`PATCH`中。
- `Timeout` 请求响应超时设置
- `Client` HTTP请求的Client，如果未指定则使用默认值：`http.DefaultCleint`
- `Adapter` 能自定义HTTP请求的处理函数，主要方便各类mock测试场景
- `RequestInterceptors` 请求的相关拦截器
- `EnableTrace` 是否启用事件跟踪，包括HTTP请求中的DNS解析、HTTP发送、开始接收数据等事件
- `OnError` 当请求出错时回调，可在此处重新对出错封装为自定义出错类型或出错率监控
- `OnDone` 请求完成时回调，包括成功或失败的请求，用于HTTP请求的相关性能与出错统计