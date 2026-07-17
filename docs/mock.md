---
description: Mock
---

# Mock

对于HTTP类的服务，测试时最重要的一点就是如何`mock`了，一般都是使用[gock](https://github.com/h2non/gock)针对http.Client来mock数据。下面我们来直接使用自带的mock方法（v2中mock通过atomic实现，并发安全）：

```go
package main

import (
	"context"
	"fmt"

	axios "github.com/vicanso/go-axios/v2"
)

type (
	// UserInfo user info
	UserInfo struct {
		Account string `json:"account,omitempty"`
		Name    string `json:"name,omitempty"`
	}
)

var (
	aslant = axios.NewClient(
		axios.WithBaseURL("https://aslant.site/"),
	)
)

// getUserInfo get user info from aslant.site
func getUserInfo(ctx context.Context) (*UserInfo, error) {
	return axios.GetJSON[UserInfo](ctx, aslant, "/users/me")
}

func main() {
	done := aslant.Mock(&axios.Response{
		Data:   []byte(`{"account":"tree", "name":"tree.xie"}`),
		Status: 200,
	})
	defer done()
	userInfo, err := getUserInfo(context.Background())
	fmt.Println(err)
	fmt.Println(userInfo)
}
```

还有`MultiMock`方法，提供一次针对多个route的mock处理，它的使用和`Mock`类似，只是参数为`map[string]*axios.Response`（key为route，默认为URL的path）。`CustomMock`则可根据config动态生成响应。
