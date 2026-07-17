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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	assert := assert.New(t)

	var nilErr *Error
	assert.Equal("", nilErr.Error())
	assert.Nil(nilErr.Unwrap())

	err := &Error{Message: "custom"}
	assert.Equal("custom", err.Error())

	err = &Error{Err: errors.New("root")}
	assert.Equal("root", err.Error())
	assert.Equal("root", err.Unwrap().Error())

	err = &Error{Status: 503}
	assert.Equal("request failed with status code 503", err.Error())

	err = &Error{}
	assert.Equal("request failed", err.Error())

	statusErr := newStatusError(&Config{URL: "/"}, &Response{Status: 400, Data: []byte("bad")})
	assert.True(errors.Is(statusErr, ErrStatusInvalid))
	assert.Equal(400, statusErr.Status)
	assert.Equal("bad", string(statusErr.Response.Data))

	bodyErr := newBodyTooLargeError(&Config{}, &Response{Data: []byte("12345")})
	assert.True(errors.Is(bodyErr, ErrBodyTooLarge))
}
