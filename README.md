# go-axios

[![Build Status](https://github.com/vicanso/go-axios/workflows/Test/badge.svg)](https://github.com/vicanso/go-axios/actions)

简单易用的HTTP客户端，参考[axios](https://github.com/axios/axios)的相关实现，支持各类不同的`interceptor`与`transform`，特性如下：

- 支持自定义的transform，可根据应用场景指定各类不同的参数格式
- 支持Request与Response的interceptor，可针对请求参数与响应数据添加各类的转换处理
- 默认支持gzip与br两种压缩方式，节约带宽占用
- 支持启用请求中的各事件记录，可细化请求的各场景耗时
- 简单易用的Mock处理
- Error与Done事件的支持，通过监听事件可快捷收集服务的出错与性能统计

```go
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vicanso/go-axios"
)

func main() {
	// 使用默认的配置
	resp, err := axios.Get("https://www.baidu.com")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Data)

	// 自定义instance，可指定client、interceptor等
	ins := axios.NewInstance(&axios.InstanceConfig{
		BaseURL:     "https://www.baidu.com",
		EnableTrace: true,
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		// 超时设置，建议设置此字段避免无接口无超时处理
		Timeout: 10 * time.Second,
		// OnDone 无论成功或失败均会调用，可在此处添加统计
		OnDone: func(config *axios.Config, resp *axios.Response, err error) {
			fmt.Println(config)
			fmt.Println(resp)
			fmt.Println(err)
		},
	})
	resp, err = ins.Get("/")
	if err != nil {
		panic(err)
	}
	buf, _ := json.Marshal(resp.Config.HTTPTrace.Stats())
	fmt.Println(resp.Config.HTTPTrace.Stats())
	fmt.Println(string(buf))
	fmt.Println(resp.Config.HTTPTrace.Protocol)
	fmt.Println(resp.Status)
	fmt.Println(string(resp.Data))
}
```
