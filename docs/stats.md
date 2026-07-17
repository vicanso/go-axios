---
description: stats性能统计
---

# EnableTrace

可以指定Client启用trace，启用之后能记录DNS、TCP连接、TLS连接以及数据接收等各事件的发生时间，生成相应的统计数据。建议将性能统计的数据写入influxdb，之后基于统计数据生成各类监控指标。

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	axios "github.com/vicanso/go-axios/v2"
)

var (
	aslant = axios.NewClient(
		axios.WithBaseURL("https://npmtrend.com"),
		axios.WithEnableTrace(),
		axios.WithResponseInterceptor(httpStats),
	)
)

func httpStats(resp *axios.Response) (err error) {
	stats := make(map[string]any)
	config := resp.Config
	stats["url"] = config.URL
	stats["status"] = resp.Status

	ht := config.HTTPTrace
	if ht != nil {
		stats["timeline"] = ht.Stats()
		stats["addr"] = ht.Addr
		stats["reused"] = ht.Reused
	}
	// 可以将相应的记录写入统计数据
	buf, _ := json.Marshal(stats)
	fmt.Println(string(buf))
	return nil
}

func main() {
	resp, err := aslant.Get(context.Background(), "/")
	fmt.Println(err)
	fmt.Println(resp.Status)
}
```

也可以在`OnDone`中通过`GetStats(config, err)`直接获取汇总后的`Stats`结构，包含route、method、状态码、各阶段耗时（ms）与响应大小。
