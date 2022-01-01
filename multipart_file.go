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
	"bytes"
	"mime/multipart"
)

type multipartFile struct {
	data   *bytes.Buffer
	writer *multipart.Writer
}

// NewMultipartFile creates a multipart file writer
func NewMultipartFile() *multipartFile {
	// 初始化容量较大的buffer
	// 用于文件上传
	buf := make([]byte, 0, 100*1024)
	b := bytes.NewBuffer(buf)
	return &multipartFile{
		data:   b,
		writer: multipart.NewWriter(b),
	}
}

// AddFile adds file to writer
func (m *multipartFile) AddFile(field, filename string, data []byte) error {
	w, err := m.writer.CreateFormFile(field, filename)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// AddFields adds fields to writer
func (m *multipartFile) AddFields(fields map[string]string) error {
	for k, v := range fields {
		err := m.writer.WriteField(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// FormDataContentType returns content type
func (m *multipartFile) FormDataContentType() string {
	return m.writer.FormDataContentType()
}

// Bytes returns bytes
func (m *multipartFile) Bytes() ([]byte, error) {
	err := m.writer.Close()
	if err != nil {
		return nil, err
	}
	return m.data.Bytes(), nil
}
