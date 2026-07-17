//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	axios "github.com/vicanso/go-axios/v2"
)

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
		axios.WithBaseURL("https://www.baidu.com"),
		axios.WithEnableTrace(),
		axios.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}),
		// 超时设置，建议设置此字段避免接口无超时处理
		axios.WithTimeout(10*time.Second),
		// OnDone 无论成功或失败均会调用，可在此处添加统计
		axios.WithOnDone(func(config *axios.Config, resp *axios.Response, err error) {
			fmt.Println(config)
			fmt.Println(resp)
			fmt.Println(err)
		}),
	)
	resp, err = ins.Get(ctx, "/")
	if err != nil {
		panic(err)
	}
	buf, _ := json.Marshal(resp.Config.HTTPTrace.Stats())
	fmt.Println(string(buf))
	fmt.Println(resp.Config.HTTPTrace.Protocol)
	fmt.Println(resp.Status)
	fmt.Println(string(resp.Data))
}
