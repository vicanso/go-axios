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
