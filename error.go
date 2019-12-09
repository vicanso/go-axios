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
	"fmt"
	"io"
	"net/url"
)

type (
	// Error error for axios
	Error struct {
		Code    int     `json:"code,omitempty"`
		Message string  `json:"message,omitempty"`
		Config  *Config `json:"-"`
		Err     error   `json:"-"`
	}
)

// Error error interface
func (e *Error) Error() string {
	str := fmt.Sprintf("message=%s", e.Err.Error())

	if e.Code != 0 {
		str = fmt.Sprintf("code=%d, %s", e.Code, str)
	}
	return str
}

// Format error format
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	default:
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

// Timeout timeout error
func (e *Error) Timeout() bool {
	if e.Err == nil {
		return false
	}
	err, ok := e.Err.(*url.Error)
	if !ok {
		return false
	}
	return err.Timeout()
}

// CreateError create an error
func CreateError(err error, config *Config, code int) *Error {
	e, ok := err.(*Error)
	// 如果已经是Error，直接返回
	if ok {
		return e
	}
	return &Error{
		Message: err.Error(),
		Code:    code,
		Config:  config,
		Err:     err,
	}
}
