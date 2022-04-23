// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package axios

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLJoin(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("https://aslant.site/users/me", urlJoin("https://aslant.site", "/users/me"))
	assert.Equal("https://aslant.site/users/me", urlJoin("https://aslant.site/", "/users/me"))
	assert.Equal("https://aslant.site/users/me", urlJoin("https://aslant.site/", "https://aslant.site/users/me"))
	assert.Equal("http://aslant.site/users/me", urlJoin("https://aslant.site/", "http://aslant.site/users/me"))
	assert.Equal("http://aslant.site/users/me", urlJoin("", "http://aslant.site/users/me"))
	assert.Equal("http://aslant.site/users/me", urlJoin("http://aslant.site/", "/users/me"))
	assert.Equal("http://aslant.site/users/me", urlJoin("http://aslant.site", "/users/me"))
	assert.Equal("http://aslant.site/users/me", urlJoin("http://aslant.site/", "users/me"))
	assert.Equal("http://aslant.site/users/me", urlJoin("http://aslant.site", "users/me"))
}

func TestIsNeedToTransformRequestBody(t *testing.T) {
	assert := assert.New(t)

	assert.True(isNeedToTransformRequestBody("POST"))
	assert.True(isNeedToTransformRequestBody("PATCH"))
	assert.True(isNeedToTransformRequestBody("PUT"))
	assert.False(isNeedToTransformRequestBody("GET"))
}

func TestConfig(t *testing.T) {
	assert := assert.New(t)
	config := &Config{}

	strValue := "1"
	bValue := true
	iValue := 2

	config.Set("s", strValue)
	config.Set("b", bValue)
	config.Set("i", iValue)

	assert.Equal(strValue, config.Get("s"))
	assert.Equal(strValue, config.GetString("s"))
	assert.Equal("", config.GetString("s1"))

	assert.Equal(bValue, config.GetBool("b"))
	assert.Equal(false, config.GetBool("b1"))

	assert.Equal(iValue, config.GetInt("i"))
	assert.Equal(0, config.GetInt("i1"))

	queryKey := "a"
	queryValue := "1"
	config.AddQuery(queryKey, queryValue)
	assert.Equal(queryValue, config.Query.Get(queryKey))

	paramKey := "c"
	paramValue := "2"
	config.AddParam(paramKey, paramValue)
	assert.Equal(paramValue, config.Params[paramKey])
}

func TestAddQuery(t *testing.T) {
	assert := assert.New(t)
	config := &Config{}
	config.AddQuery("a", "1")
	assert.Equal("1", config.Query.Get("a"))
}

func TestAddQueryMap(t *testing.T) {
	assert := assert.New(t)
	config := &Config{
		URL: "/",
	}
	config.AddQueryMap(map[string]string{
		"a": "1",
		"b": "2",
	})
	assert.Equal("/?a=1&b=2", config.GetURL())
}

func TestAddQueryStruct(t *testing.T) {
	assert := assert.New(t)
	config := &Config{
		URL: "/",
	}
	type Data struct {
		Count       int     `json:"count,omitempty"`
		Number      uint    `json:"number,omitempty"`
		IsVIP       bool    `json:"isVIP,omitempty"`
		Name        string  `json:"name,omitempty"`
		Amount      float64 `json:"amount,omitempty"`
		Category    string  `json:"category,omitempty"`
		IgnoreField string  `json:"-"`
	}
	_, err := config.AddQueryStruct(&Data{
		Count:       1,
		Number:      2,
		IsVIP:       true,
		Name:        "test",
		Amount:      10.2,
		IgnoreField: "aaaaaaaa",
	})
	assert.Nil(err)
	assert.Equal("/?amount=10.200&count=1&isVIP=true&name=test&number=2", config.GetURL())
}

func TestAddParam(t *testing.T) {
	assert := assert.New(t)
	config := &Config{}
	config.AddParam("a", "1")
	assert.Equal("1", config.Params["a"])
}

func TestGetRequestBody(t *testing.T) {
	assert := assert.New(t)

	r := bytes.NewBufferString("abc")
	config := &Config{
		Method: "POST",
		Body:   r,
	}
	body, err := config.getRequestBody()
	assert.Nil(err)
	assert.Equal(r, body)
}

func TestCURL(t *testing.T) {
	assert := assert.New(t)
	query := make(url.Values)
	query.Add("a", "1")
	query.Add("a", "2")
	conf := Config{
		BaseURL:          "http://test.com",
		Headers:          make(http.Header),
		TransformRequest: DefaultTransformRequest,
		URL:              "/users/:type",
		Params: map[string]string{
			"type": "vip",
		},
		Query:  query,
		Method: "POST",
		Body: &struct {
			Name  string `json:"name,omitempty"`
			Count int    `json:"count,omitempty"`
		}{
			"nickname",
			10,
		},
	}

	assert.Equal(`curl -XPOST -d '{"name":"nickname","count":10}' -H 'Content-Type:application/json;charset=utf-8' 'http://test.com/users/vip?a=1&a=2'`, conf.CURL())

	data := make(url.Values)
	data.Add("name", "nickname")
	data.Add("count", "10")
	conf = Config{
		BaseURL:          "http://test.com",
		Headers:          make(http.Header),
		TransformRequest: DefaultTransformRequest,
		URL:              "/users/:type",
		Params: map[string]string{
			"type": "vip",
		},
		Query:  query,
		Method: "POST",
		Body:   data,
	}

	assert.Equal(`curl -XPOST -d 'count=10&name=nickname' -H 'Content-Type:application/x-www-form-urlencoded;charset=utf-8' 'http://test.com/users/vip?a=1&a=2'`, conf.CURL())

	conf = Config{
		BaseURL:          "http://test.com",
		Headers:          make(http.Header),
		TransformRequest: DefaultTransformRequest,
		URL:              "/users/:type",
		Params: map[string]string{
			"type": "vip",
		},
		Query: query,
	}
	assert.Equal(`curl -XGET 'http://test.com/users/vip?a=1&a=2'`, conf.CURL())

}

func TestBaseConfig(t *testing.T) {
	assert := assert.New(t)

	bc := baseConfig{}

	assert.Empty(bc.onBeforeNewRequests)
	// add
	bc.AddBeforeNewRequestListener(func(config *Config) (err error) {
		return nil
	})
	assert.Equal(1, len(bc.onBeforeNewRequests))

	onBeforeNewRequests := bc.onBeforeNewRequests
	// prepend
	bc.PrependBeforeNewRequestListener(func(config *Config) (err error) {
		return nil
	})
	assert.Equal(1, len(onBeforeNewRequests))
	assert.Equal(2, len(bc.onBeforeNewRequests))

	assert.Empty(bc.onErrors)
	bc.AddErrorListener(func(err error, config *Config) (newErr error) {
		return nil
	})
	assert.Equal(1, len(bc.onErrors))
	onErrors := bc.onErrors
	bc.PrependErrorListener(func(err error, config *Config) (newErr error) {
		return nil
	})
	assert.Equal(1, len(onErrors))
	assert.Equal(2, len(bc.onErrors))

	assert.Empty(bc.onDones)
	bc.AddDoneListener(func(config *Config, resp *Response, err error) {
	})
	assert.Equal(1, len(bc.onDones))
	onDones := bc.onDones
	bc.PrependDoneListener(func(config *Config, resp *Response, err error) {
	})
	assert.Equal(1, len(onDones))
	assert.Equal(2, len(bc.onDones))
}
