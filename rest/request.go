package rest

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/go-chi/chi"
	"github.com/gonzispina/gokit/context"
	"github.com/gonzispina/gokit/errors"
	"github.com/gonzispina/gokit/logs"
)

// ContentType pse
type ContentType string

// String casts
func (c ContentType) String() string {
	return string(c)
}

// ContentTypes array
type ContentTypes []ContentType

// Has content type
func (c ContentTypes) Has(contentType string) bool {
	for _, ct := range c {
		if ct.String() == contentType {
			return true
		}
	}
	return false
}

// String cast
func (c ContentTypes) String() string {
	var res []string
	for _, ct := range c {
		res = append(res, ct.String())
	}
	return strings.Join(res, ", ")
}

const (
	// ApplicationJSON content type
	ApplicationJSON ContentType = "application/json"
	// ApplicationPDF content type
	ApplicationPDF ContentType = "application/pdf"
	// ImagePNG content type
	ImagePNG ContentType = "image/png"
	// ImageJPEG content type
	ImageJPEG ContentType = "image/jpeg"
	// ImageJPG content type
	ImageJPG ContentType = "image/jpg"
	// AudioMP4A content type
	AudioMP4A ContentType = "audio/mp4"
	// ApplicationDocx content type
	ApplicationDocx ContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	// ApplicationDoc content type
	ApplicationDoc ContentType = "application/msword"
	// ApplicationTxt content type
	ApplicationTxt ContentType = "text/plain"
)

// HandlerFunc implementation to isolate the web entry points from the framework
// and make entry points more "testable"
type HandlerFunc func(r *Request) *Response

// NewResponse for all entry points
func NewResponse(statusCode int, data interface{}, header map[string]string) *Response {
	if header == nil {
		header = map[string]string{}
	}
	if data != nil && header[ContentTypeHeader] == "" {
		header[ContentTypeHeader] = string(ApplicationJSON)
	}
	return &Response{
		Data:       data,
		StatusCode: statusCode,
		Header:     header,
	}
}

// Response implementation
type Response struct {
	Data       interface{}       `json:"-"`
	StatusCode int               `json:"-"`
	Header     map[string]string `json:"-"`
	Err        string            `json:"description"`
	Code       string            `json:"code"`
}

// File implementation
type File struct {
	Name        string
	Ext         string
	Data        io.Reader
	ContentType ContentType
	Size        int64
}

// Request implementation
type Request struct {
	*http.Request
	UserID      string
	IPAddress   string
	RouteParams map[string]string
	queryParams map[string]string
	Body        io.Reader
	File        *File
	ctx         context.Context
	JSONBody    interface{}
	Filter      *Filter
}

// Context of the request
func (r *Request) Context() context.Context {
	if r.ctx == nil {
		r.ctx = context.Background()
	}
	return r.ctx
}

// URLParam returns a param from the URL
func (r *Request) URLParam(key string) string {
	return r.RouteParams[key]
}

// SetURLParam adds an URL param to the request
func (r *Request) SetURLParam(key, value string) *Request {
	if r.RouteParams == nil {
		r.RouteParams = map[string]string{}
	}
	r.RouteParams[key] = value
	return r
}

// QueryParam return a query param from the URL
func (r *Request) QueryParam(key string) string {
	// This is done like this to test easily
	value, found := r.queryParams[key]
	if found {
		return value
	}
	if r.Request == nil {
		return ""
	}
	r.queryParams[key] = r.URL.Query().Get(key)
	return r.queryParams[key]
}

// SetQueryParam adds a query param to the request
func (r *Request) SetQueryParam(key, value string) *Request {
	if r.queryParams == nil {
		r.queryParams = map[string]string{}
	}
	r.queryParams[key] = value
	return r
}

// UpgradeMiddleware transforms a chi/http request into Request
func UpgradeMiddleware(logger logs.Logger) func(handlerFunc HandlerFunc) http.HandlerFunc {
	if logger == nil {
		panic("logger must be initialized")
	}
	return func(handler HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rctx := chi.RouteContext(r.Context())
			ctx := context.Upgrade(r.Context())

			defer func() {
				r := recover()
				if r == nil {
					return
				}

				logger.Error(ctx, "Recovered from panic", logs.Error(r.(error)), logs.Bytes(debug.Stack()))
				w.WriteHeader(http.StatusInternalServerError)
			}()

			userID := r.Header.Get(CallerIDHeader)
			routeParams := map[string]string{}
			for i, key := range rctx.URLParams.Keys {
				routeParams[key] = rctx.URLParams.Values[i]
			}

			ipAddress := r.RemoteAddr
			if strings.Index(ipAddress, ":") > 0 {
				host, _, err := net.SplitHostPort(ipAddress)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				ipAddress = host
			}

			req := &Request{
				UserID:      userID,
				Body:        r.Body,
				RouteParams: routeParams,
				queryParams: map[string]string{},
				ctx:         ctx,
				Filter:      &Filter{},
				IPAddress:   ipAddress,
				Request:     r,
			}

			res := handler(req)
			if res.Header != nil {
				for k, v := range res.Header {
					w.Header().Add(k, v)
				}
			}

			if res.Err != "" {
				w.Header().Add(ContentTypeHeader, ApplicationJSON.String())
				w.WriteHeader(res.StatusCode)
				_ = json.NewEncoder(w).Encode(res)
				return
			}

			if res.Data == nil {
				w.WriteHeader(res.StatusCode)
				return
			}

			if res.Header[ContentTypeHeader] == string(ApplicationJSON) || res.Header[ContentTypeHeader] == "" {
				w.Header().Add(ContentTypeHeader, ApplicationJSON.String())
				w.WriteHeader(res.StatusCode)
				err := json.NewEncoder(w).Encode(&res.Data)
				if err != nil {
					logger.Error(ctx, "Couldn't marshal response", logs.Error(err), logs.UserID(userID))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			}

			data, ok := res.Data.(io.ReadCloser)
			if !ok {
				logger.Error(ctx, "Cannot handle datatype", zap.String("type", reflect.TypeOf(res.Data).String()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer func() {
				if err := data.Close(); err != nil {
					logger.Error(ctx, "Couldn't close reader", logs.Error(err))
				}
			}()

			w.Header().Add(ContentTypeHeader, res.Header[ContentTypeHeader])
			w.WriteHeader(res.StatusCode)
			_, err := io.Copy(w, data)
			if err != nil {
				logger.Error(ctx, "Couldn't write Data into response", logs.Error(err))
				return
			}
		}
	}
}

// JSONMiddleware validates that a request has a JSON body type
// It returns 400 Bad request if it cannot parse the JSON into the provided struct tag
func JSONMiddleware(logger logs.Logger) func(handler HandlerFunc, t reflect.Type, validator *Validator) HandlerFunc {
	return func(handler HandlerFunc, t reflect.Type, validator *Validator) HandlerFunc {
		if t.Kind() == reflect.Ptr {
			panic(fmt.Sprintf("Concept %s cannot be a pointer", t.Elem()))
		}
		return func(r *Request) *Response {
			if r.Method == http.MethodGet || r.Method == http.MethodDelete {
				return handler(r)
			}

			contentType := r.Header.Get(ContentTypeHeader)
			if contentType != "application/json" {
				return BadRequest(errors.New("invalid content type", "invalid_content_type"))
			}

			value := reflect.New(t)
			err := json.NewDecoder(r.Body).Decode(value.Interface())
			if err != nil {
				message := "invalid request body"
				var jsonErr *json.UnmarshalTypeError
				if errors.As(err, &jsonErr) {
					message = fmt.Sprintf("field '%s' cannot be of type '%s'. It must be of type '%s'", jsonErr.Field, jsonErr.Value, jsonErr.Type.String())
				}
				return BadRequest(errors.New(message, "invalid_request_body"))
			}

			if err = validator.Value(value); err != nil {
				if err == ErrLib {
					logger.Error(r.Context(), "An error occurred in validator lib", logs.Error(err))
					return InternalServerError()
				}
				return BadRequest(err)
			}

			r.JSONBody = value.Interface()

			return handler(r)
		}
	}
}

// UploadMiddleware validates that a request has a a valid content type and length
// It returns 408 Entity Too Large if it the content length is not valid or 400 Bad Request if
// the content type is not valid
func UploadMiddleware(logger logs.Logger) func(handler HandlerFunc, maxFileSize int64, t reflect.Type, validator *Validator, types ...ContentType) HandlerFunc {
	return func(handler HandlerFunc, maxFileSize int64, t reflect.Type, validator *Validator, types ...ContentType) HandlerFunc {
		if len(types) == 0 {
			panic("content type array cannot be empty")
		}
		if maxFileSize < 1 {
			panic("max file size must be > 0")
		}
		if t != nil && t.Kind() == reflect.Ptr {
			panic(fmt.Sprintf("Concept %s cannot be a pointer", t.Elem()))
		}

		contentTypes := ContentTypes(types)
		return func(r *Request) *Response {
			if r.Method != http.MethodPost {
				return handler(r)
			}

			err := r.ParseMultipartForm(maxFileSize << 20)
			if err != nil {
				if errors.Is(err, multipart.ErrMessageTooLarge) {
					return RequestEntityTooLarge()
				}
				return BadRequest(errors.New("invalid multipart body", "invalid_multipart_body"))
			}

			data, header, err := r.FormFile("data")
			if err != nil {
				logger.Warn(r.Context(), "Couldn't get data", logs.Error(err))
				return BadRequest(errors.New("'data' form value must be present", "multipart_field_data_not_present"))
			}

			contentType := header.Header.Get(ContentTypeHeader)
			if !contentTypes.Has(contentType) {
				return BadRequest(errors.New("content type must be one of: "+contentTypes.String(), "invalid_content_type"))
			}

			r.File = &File{
				Name:        header.Filename,
				Size:        header.Size,
				Ext:         filepath.Ext(header.Filename),
				Data:        data,
				ContentType: ContentType(contentType),
			}

			if t != nil {
				value := reflect.New(t)
				str := r.FormValue("params")
				if len(str) != 0 {
					err = json.Unmarshal([]byte(str), value.Interface())
					if err != nil {
						logger.Warn(r.Context(), "Invalid body params")
						return BadRequest(errors.New("'params' must be a valid json", "multipart_invalid_field_params"))
					}
				} else {
					params, header, err := r.FormFile("params")
					if err != nil {
						logger.Warn(r.Context(), "Invalid body params")
						return BadRequest(errors.New("'params' field invalid", "multipart_invalid_field_params"))
					}

					contentType := header.Header.Get(ContentTypeHeader)
					if contentType != string(ApplicationJSON) {
						return BadRequest(errors.New("'params' content type invalid", "multipart_invalid_field_params"))
					}

					err = json.NewDecoder(params).Decode(value.Interface())
					if err != nil {
						return BadRequest(errors.New("'params' must be a valid json", "multipart_invalid_field_params"))
					}
				}

				if err = validator.Value(value); err != nil {
					if err == ErrLib {
						logger.Error(r.Context(), "An error occurred in validator lib", logs.Error(err))
						return InternalServerError()
					}
					return BadRequest(err)
				}

				r.JSONBody = value.Interface()
			}

			return handler(r)
		}
	}
}
