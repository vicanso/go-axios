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
	"fmt"
)

var (
	// ErrBodyTooLarge is returned when the response body exceeds MaxBodySize.
	ErrBodyTooLarge = errors.New("response body exceeds max size")
	// ErrStatusInvalid is returned when ValidateStatus rejects the status code.
	// Prefer errors.As to *Error for status and response details.
	ErrStatusInvalid = errors.New("request failed with invalid status code")
)

// Error is a structured request error. It may wrap a lower-level error and
// optionally carry the HTTP status and response when available.
type Error struct {
	// Message is an optional human-readable description.
	Message string
	// Status is the HTTP status code when a response was received.
	Status int
	// Response is the response associated with the error, if any.
	Response *Response
	// Config is the request config associated with the error, if any.
	Config *Config
	// Err is the underlying error, if any.
	Err error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.Status > 0 {
		return fmt.Sprintf("request failed with status code %d", e.Status)
	}
	return "request failed"
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func newStatusError(config *Config, resp *Response) *Error {
	status := 0
	if resp != nil {
		status = resp.Status
	}
	return &Error{
		Message:  fmt.Sprintf("request failed with status code %d", status),
		Status:   status,
		Response: resp,
		Config:   config,
		Err:      ErrStatusInvalid,
	}
}

func newBodyTooLargeError(config *Config, resp *Response) *Error {
	return &Error{
		Message:  ErrBodyTooLarge.Error(),
		Response: resp,
		Config:   config,
		Err:      ErrBodyTooLarge,
	}
}
