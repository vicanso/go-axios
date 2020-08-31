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
