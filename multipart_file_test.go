// Copyright 2021 tree xie
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

func TestMultipartFile(t *testing.T) {
	assert := assert.New(t)

	mf := NewMultipartFile()
	mf.writer.SetBoundary("abc")

	err := mf.AddFile("file", "test.go", []byte("Hello World!"))
	assert.Nil(err)

	err = mf.AddFields(map[string]string{
		"type": "vip",
	})
	assert.Nil(err)

	assert.Equal("multipart/form-data; boundary=abc", mf.FormDataContentType())

	data, err := mf.Bytes()
	assert.Nil(err)
	assert.Equal("--abc\r\nContent-Disposition: form-data; name=\"file\"; filename=\"test.go\"\r\nContent-Type: application/octet-stream\r\n\r\nHello World!\r\n--abc\r\nContent-Disposition: form-data; name=\"type\"\r\n\r\nvip\r\n--abc--\r\n", string(data))
}
