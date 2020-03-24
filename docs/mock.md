---
description: Mock
---

# Mock

对于HTTP类的服务，测试时最重要的一点就是如何`mock`了，一般都是使用[gock](https://github.com/h2non/gock)针对http.Client来mock数据。下面我们来直接使用自带的mock方法：

```go
package main

import (
	"fmt"

	"github.com/vicanso/go-axios"
)

type (
	// UserInfo user info
	UserInfo struct {
		Account string `json:"account,omitempty"`
		Name    string `json:"name,omitempty"`
	}
)

var (
	aslant = axios.NewInstance(&axios.InstanceConfig{
		BaseURL: "https://aslant.site/",
	})
)

// getUserInfo get user info from aslant.site
func getUserInfo() (userInfo *UserInfo, err error) {
	resp, err := aslant.Get("/users/me")
	if err != nil {
		return
	}
	userInfo = new(UserInfo)
	err = resp.JSON(userInfo)
	if err != nil {
		return
	}
	return
}

func main() {
	done := aslant.Mock(&axios.Response{
		Data: []byte(`{"account":"tree", "name":"tree.xie"}`),
		Status: 200,
	});
	defer done()
	userInfo, err := getUserInfo()
	fmt.Println(err)
	fmt.Println(userInfo)
}
```

还有MultiMock方法，提供一次针对多个path的mock处理，它的使用和Mock类似，只是参数为map[string]*axios.Response。
