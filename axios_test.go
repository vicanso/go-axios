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

func TestDefaultInstance(t *testing.T) {
	assert := assert.New(t)
	ins := GetDefaultInstance()
	original := ins.Config.Adapter
	defer func() {
		ins.Config.Adapter = original
	}()
	ins.Config.Adapter = func(conf *Config) (*Response, error) {
		return &Response{
			Config: conf,
		}, nil
	}
	resp, err := Request(&Config{})
	assert.Nil(err)
	assert.Equal("GET", resp.Config.Method)

	resp, err = Get("/")
	assert.Nil(err)
	assert.Equal("GET", resp.Config.Method)

	resp, err = Delete("/")
	assert.Nil(err)
	assert.Equal("DELETE", resp.Config.Method)

	resp, err = Head("/")
	assert.Nil(err)
	assert.Equal("HEAD", resp.Config.Method)

	resp, err = Options("/")
	assert.Nil(err)
	assert.Equal("OPTIONS", resp.Config.Method)

	resp, err = Post("/", nil)
	assert.Nil(err)
	assert.Equal("POST", resp.Config.Method)

	resp, err = Put("/", nil)
	assert.Nil(err)
	assert.Equal("PUT", resp.Config.Method)

	resp, err = Patch("/", nil)
	assert.Nil(err)
	assert.Equal("PATCH", resp.Config.Method)
}
