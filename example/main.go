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
		EnableTrace: true,
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		Timeout: 10 * time.Second,
	})
	resp, err = ins.Get("https://www.baidu.com/")
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
