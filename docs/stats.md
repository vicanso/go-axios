---
description: stats性能统计
---

# EnableTrace

可以指定Instance启用trace，启用之后能记录DNS、TCP连接、TLS连接以及数据接收等各事件的发生时间，生成相应的统计数据。 建议将性能统计的数据写入influxdb，之后基本统计数据生成各类监控指标。

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/vicanso/go-axios"
)

var (
	aslant = axios.NewInstance(&axios.InstanceConfig{
		BaseURL:     "https://npmtrend.com",
		EnableTrace: true,
		ResponseInterceptors: []axios.ResponseInterceptor{
			httpStats,
		},
	})
)

func httpStats(resp *axios.Response) (err error) {
	stats := make(map[string]interface{})
	config := resp.Config
	stats["url"] = config.URL
	stats["status"] = resp.Status

	ht := config.HTTPTrace
	if ht != nil {
		stats["timeline"] = config.HTTPTrace.Stats()
		stats["addr"] = ht.Addr
		stats["reused"] = ht.Reused
	}
	// 可以将相应的记录写入统计数据
	buf, _ := json.Marshal(stats)
	fmt.Println(string(buf))
	return nil
}

func main() {
	resp, err := aslant.Get("/")
	fmt.Println(err)
	fmt.Println(resp.Status)
}

```