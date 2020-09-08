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
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		Query:  query,
		Method: "GET",
	}
	assert.Equal(`curl -XGET 'http://test.com/users/vip?a=1&a=2'`, conf.CURL())

}
