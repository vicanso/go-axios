package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/vicanso/go-axios"
)

func main() {
	ins := axios.NewInstance(&axios.InstanceConfig{
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		Timeout: 10 * time.Second,
	})
	resp, err := ins.Get("https://www.baidu.com/")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp.Data))
}
