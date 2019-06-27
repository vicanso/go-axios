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
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

// CreateError create an error
func CreateError(err error, config *Config, code int) *Error {
	return &Error{
		Message: err.Error(),
		Code:    code,
		Config:  config,
		Err:     err,
	}
}
