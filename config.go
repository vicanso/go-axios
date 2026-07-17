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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	HT "github.com/vicanso/http-trace"
)

type (
	// RequestInterceptor is called after the http request is created
	// and before it is sent, it can modify the request.
	RequestInterceptor func(config *Config) (err error)
	// ResponseInterceptor is called after the response is received
	// and transformed, it can modify the response.
	ResponseInterceptor func(resp *Response) (err error)
	// OnError is called when the request fails, it can replace the
	// error by returning a new one.
	OnError func(err error, config *Config) (newErr error)
	// OnDone is called when the request is done, whether it succeeds or fails.
	OnDone func(config *Config, resp *Response, err error)
	// OnBeforeRequest is called before the http request is created.
	// If it sets config.Response, the request will not be sent and
	// the response is returned directly.
	OnBeforeRequest func(config *Config) (err error)
	// ValidateStatus decides whether a response status code is considered
	// successful. When it returns false, the request fails with *Error.
	// If nil, all status codes are accepted (default).
	ValidateStatus func(status int) bool

	// Config is the config of a single http request.
	Config struct {
		Request  *http.Request
		Response *Response
		// Route the request route, default is the path of URL
		Route string
		// URL the request url
		URL string
		// Method http request method, default is `GET`
		Method string
		// BaseURL http request base url
		BaseURL string
		// TransformRequest transform request body
		TransformRequest []TransformRequest
		// TransformResponse transform response body
		TransformResponse []TransformResponse
		// Headers custom headers for request
		Headers http.Header
		// Params params for the request route, a path segment of the form
		// `:key` will be replaced by the url path-escaped value
		Params map[string]string
		// Query query for request
		Query url.Values

		// Body the request body
		Body any

		// Concurrency the amount of handling request of the client
		// when this request is sent
		Concurrency int32

		// Timeout request timeout
		Timeout time.Duration

		// MaxBodySize is the maximum response body size in bytes.
		// Zero means unlimited.
		MaxBodySize int64

		// ValidateStatus validates response status codes.
		// If nil, all status codes are accepted.
		ValidateStatus ValidateStatus

		// Retry is the retry policy for this request.
		// If nil, the client retry policy is used (if any).
		Retry *RetryConfig

		// Attempt is the 1-based attempt number of the last try.
		// It is set by the client and is useful in OnDone / interceptors.
		Attempt int

		// HTTPClient http client to send the request
		HTTPClient *http.Client
		// Adapter custom handling of request
		Adapter Adapter
		// RequestInterceptors request interceptor list
		RequestInterceptors []RequestInterceptor
		// ResponseInterceptors response interceptor list
		ResponseInterceptors []ResponseInterceptor

		// OnError error listener list
		OnError []OnError
		// OnDone done listener list
		OnDone []OnDone
		// OnBeforeRequest before request listener list
		OnBeforeRequest []OnBeforeRequest

		// HTTPTrace the trace info of request,
		// it's not nil when trace is enabled
		HTTPTrace *HT.HTTPTrace

		enableTrace   bool
		ctx           context.Context
		jsonMarshal   JSONMarshal
		jsonUnmarshal JSONUnmarshal
		values        map[string]any
	}

	// RequestOption sets an option of request config
	RequestOption func(config *Config)
)

var ErrRequestDataTypeInvalid = errors.New("request data type is not supported")
var ErrRequestIsForbidden = errors.New("request is forbidden")

// Query returns a request option that adds a query value
func Query(key, value string) RequestOption {
	return func(config *Config) {
		config.AddQuery(key, value)
	}
}

// QueryValues returns a request option that adds all query values
func QueryValues(query url.Values) RequestOption {
	return func(config *Config) {
		for key, values := range query {
			for _, value := range values {
				config.AddQuery(key, value)
			}
		}
	}
}

// QueryMap returns a request option that adds all values of map to query
func QueryMap(query map[string]string) RequestOption {
	return func(config *Config) {
		config.AddQueryMap(query)
	}
}

// Param returns a request option that sets a route param,
// `:key` in url will be replaced by the url escaped value
func Param(key, value string) RequestOption {
	return func(config *Config) {
		config.AddParam(key, value)
	}
}

// Header returns a request option that adds a http request header
func Header(key, value string) RequestOption {
	return func(config *Config) {
		if config.Headers == nil {
			config.Headers = make(http.Header)
		}
		config.Headers.Add(key, value)
	}
}

// Timeout returns a request option that sets the timeout of request
func Timeout(d time.Duration) RequestOption {
	return func(config *Config) {
		config.Timeout = d
	}
}

// Route returns a request option that sets the route of request,
// it's used for stats and multi mock
func Route(route string) RequestOption {
	return func(config *Config) {
		config.Route = route
	}
}

// Body returns a request option that sets the request body
func Body(body any) RequestOption {
	return func(config *Config) {
		config.Body = body
	}
}

// EnableTrace returns a request option that enables http trace of request
func EnableTrace() RequestOption {
	return func(config *Config) {
		config.enableTrace = true
	}
}

// MaxBodySize returns a request option that sets the max response body size.
// Zero means unlimited.
func MaxBodySize(n int64) RequestOption {
	return func(config *Config) {
		config.MaxBodySize = n
	}
}

// StatusValidator returns a request option that sets ValidateStatus.
func StatusValidator(fn ValidateStatus) RequestOption {
	return func(config *Config) {
		config.ValidateStatus = fn
	}
}

// ValidateStatus2xx accepts only 2xx status codes.
func ValidateStatus2xx(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

// RetryPolicy returns a request option that sets the retry policy.
// See RetryConfig for body replay rules and defaults.
func RetryPolicy(cfg RetryConfig) RequestOption {
	return func(config *Config) {
		cp := cfg
		config.Retry = &cp
	}
}

// Context returns the context of request config
func (conf *Config) Context() context.Context {
	if conf.ctx == nil {
		return context.Background()
	}
	return conf.ctx
}

// Get gets value from config
func (conf *Config) Get(key string) any {
	if conf.values == nil {
		return nil
	}
	return conf.values[key]
}

// Set sets value to config, it's a simple way to pass value
// between interceptors and listeners
func (conf *Config) Set(key string, value any) {
	if conf.values == nil {
		conf.values = make(map[string]any)
	}
	conf.values[key] = value
}

// GetString gets string value from config
func (conf *Config) GetString(key string) string {
	str, _ := conf.Get(key).(string)
	return str
}

// GetBool gets bool value from config
func (conf *Config) GetBool(key string) bool {
	b, _ := conf.Get(key).(bool)
	return b
}

// GetInt gets int value from config
func (conf *Config) GetInt(key string) int {
	i, _ := conf.Get(key).(int)
	return i
}

func getTag(sf reflect.StructField) string {
	tag := sf.Tag.Get("json")
	if tag != "" {
		arr := strings.Split(tag, ",")
		tag = arr[0]
	}
	if tag != "" {
		return tag
	}
	return sf.Name
}

func structToMapString(value any) (map[string]string, error) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	t := v.Type()

	m := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		tag := getTag(sf)
		// 对于 - 忽略
		if tag == "-" {
			continue
		}
		value := ""
		switch sf.Type.Kind() {
		case reflect.Bool:
			if v.Field(i).Bool() {
				value = "true"
			} else {
				value = "false"
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value = strconv.Itoa(int(v.Field(i).Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			value = strconv.Itoa(int(v.Field(i).Uint()))
		case reflect.Float32, reflect.Float64:
			value = fmt.Sprintf("%.3f", v.Field(i).Float())
		case reflect.String:
			value = v.Field(i).String()
		default:
			return nil, errors.New("unsupported type")
		}
		if value != "" {
			m[tag] = value
		}
	}
	return m, nil
}

// AddQueryMap converts map and adds to query
func (conf *Config) AddQueryMap(query map[string]string) *Config {
	for key, value := range query {
		conf.AddQuery(key, value)
	}
	return conf
}

// AddQueryStruct converts struct and adds to query
func (conf *Config) AddQueryStruct(value any) (*Config, error) {
	m, err := structToMapString(value)
	if err != nil {
		return conf, err
	}
	conf.AddQueryMap(m)
	return conf, nil
}

// AddQuery adds query to config
func (conf *Config) AddQuery(key, value string) *Config {
	if conf.Query == nil {
		conf.Query = make(url.Values)
	}
	conf.Query.Add(key, value)
	return conf
}

// AddParam adds route param to config
func (conf *Config) AddParam(key, value string) *Config {
	if conf.Params == nil {
		conf.Params = make(map[string]string)
	}
	conf.Params[key] = value
	return conf
}

func urlJoin(basicURL, url string) string {
	if basicURL == "" ||
		strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") {
		return url
	}
	if strings.HasSuffix(basicURL, "/") && strings.HasPrefix(url, "/") {
		return basicURL + url[1:]
	}
	if !strings.HasSuffix(basicURL, "/") && !strings.HasPrefix(url, "/") {
		return basicURL + "/" + url
	}
	return basicURL + url
}

// GetURL generates the url of request config
func (conf *Config) GetURL() string {
	requestURL := urlJoin(conf.BaseURL, conf.URL)
	if conf.Params != nil {
		requestURL = replacePathParams(requestURL, conf.Params)
	}

	if conf.Query != nil {
		if strings.Contains(requestURL, "?") {
			requestURL += ("&" + conf.Query.Encode())
		} else {
			requestURL += ("?" + conf.Query.Encode())
		}
	}
	return requestURL
}

// replacePathParams replaces whole path segments of the form `:key`.
// Only the path portion is modified; query and fragment are left unchanged.
// Whole-segment matching avoids prefix clashes such as `:id` vs `:idExtra`.
func replacePathParams(rawURL string, params map[string]string) string {
	if len(params) == 0 {
		return rawURL
	}

	pathPart, suffix := splitURLPath(rawURL)
	segments := strings.Split(pathPart, "/")
	for i, seg := range segments {
		if seg == "" || seg[0] != ':' {
			continue
		}
		name := seg[1:]
		if value, ok := params[name]; ok {
			segments[i] = url.PathEscape(value)
		}
	}
	return strings.Join(segments, "/") + suffix
}

// splitURLPath splits rawURL into path (scheme/host/path) and the query/fragment suffix.
func splitURLPath(rawURL string) (pathPart, suffix string) {
	for i := 0; i < len(rawURL); i++ {
		switch rawURL[i] {
		case '?', '#':
			return rawURL[:i], rawURL[i:]
		}
	}
	return rawURL, ""
}

// getRequestBody gets the request body,
// the body will be converted by transform request functions
func (conf *Config) getRequestBody() (r io.Reader, err error) {
	if conf.Body == nil {
		return
	}
	data := conf.Body
	for _, fn := range conf.TransformRequest {
		buf, e := fn(conf, data)
		if e != nil {
			err = e
			return
		}
		data = buf
	}
	r, ok := data.(io.Reader)
	if ok {
		return r, nil
	}
	buf, ok := data.([]byte)
	if !ok {
		err = ErrRequestDataTypeInvalid
		return
	}
	r = bytes.NewReader(buf)
	return
}

// CURL converts config to curl command
func (conf *Config) CURL() string {
	builder := new(strings.Builder)
	method := conf.Method
	if method == "" {
		method = http.MethodGet
	}
	fmt.Fprintf(builder, "curl -X%s ", method)

	r, _ := conf.getRequestBody()
	if r != nil {
		buf, _ := io.ReadAll(r)
		fmt.Fprintf(builder, `-d '%s' `, strings.ReplaceAll(string(buf), "'", `'\''`))
	}

	for key, values := range conf.Headers {
		for _, value := range values {
			fmt.Fprintf(builder, `-H '%s:%s' `, key, value)
		}
	}

	fmt.Fprintf(builder, `'%s'`, conf.GetURL())

	return builder.String()
}
