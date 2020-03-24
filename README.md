# go-axios

[![Build Status](https://img.shields.io/travis/vicanso/go-axios.svg?label=linux+build)](https://travis-ci.org/vicanso/go-axios)

简单易用的HTTP客户端，参考[axios](https://github.com/axios/axios)的相关实现，支持各类不同的`interceptor`与`transform`。

```go
package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/vicanso/go-axios"
	"github.com/vicanso/hes"
)

type (
	loginTokenResp struct {
		Token string `json:"token,omitempty"`
	}
)

func main() {
	ins := axios.NewInstance(&axios.InstanceConfig{
		BaseURL: "https://tiny.npmtrend.com",
		// 对于>=400的出错请求，根据响应数据转换为对应的出错
		ResponseInterceptors: []axios.ResponseInterceptor{
			func(resp *axios.Response) error {
				if resp.Status < http.StatusBadRequest {
					return nil
				}
				he := &hes.Error{}
				err := resp.JSON(he)
				// 如果返回数据非json
				if err != nil {
					he = hes.NewWithErrorStatusCode(errors.New(string(resp.Data)), resp.Status)
				}
				if he.Message == "" {
					he.Message = "未知异常"
				}
				return he
			},
		},
	})
	resp, err := ins.Get("/users/v1/me/login")
	if err != nil {
		panic(err)
	}
	tokenResp := new(loginTokenResp)
	err = resp.JSON(tokenResp)
	if err != nil {
		panic(err)
	}
	resp, err = ins.Post("/users/v1/me/login", map[string]string{
		"account": "tree.xie",
		// 密码需要通过tokenResp.Token与密码迦后传输
		"password": "md5(password + token)",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp.Data))
}
```
