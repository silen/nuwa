package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/valyala/fasthttp"

	"github.com/silen/nuwa/internal/jsonutil"
	"github.com/silen/nuwa/logs"
)

// NewHTTP ...
func NewHTTP(ctx context.Context) *HTTP {
	return &HTTP{
		Ctx: ctx,
	}
}

// HTTP ..
type HTTP struct {
	header      map[string]string
	contentType string
	Ctx         context.Context
	timeout     time.Duration
}

const defaultTimeout = 30 * time.Second

const (
	sendMethodGet        = "GET"
	sendMethodPost       = "POST"
	sendMethodJSONBody   = "JsonBody"
	sendMethodDeleteBody = "DELETEBody"

	httpMethodGet    = "GET"
	httpMethodPost   = "POST"
	httpMethodDelete = "DELETE"
)

type requestSpec struct {
	method      string
	url         string
	body        []byte
	contentType string
	logPayload  string
}

// Get ..
func (h *HTTP) SetHeader(data map[string]string) *HTTP {
	if data == nil {
		h.header = nil
		return h
	}

	h.header = make(map[string]string, len(data))
	for k, v := range data {
		h.header[k] = v
	}
	return h
}

// Get ..
func (h *HTTP) SetContentType(data string) *HTTP {
	h.contentType = data
	return h
}

func (h *HTTP) SetTimeout(timeout time.Duration) *HTTP {
	h.timeout = timeout
	return h
}

// Get ..
func (h *HTTP) Get(url string, data any, structObject any) (res string, err error) {
	return h.sendAndDecode(sendMethodGet, url, normalizeData(data), structObject)
}

// Post ..
func (h *HTTP) Post(url string, data any, structObject any) (res string, err error) {
	return h.sendAndDecode(sendMethodPost, url, normalizeData(data), structObject)
}

// Post ..
func (h *HTTP) PostJson(url string, data any, structObject any) (res string, err error) {
	return h.sendAndDecode(sendMethodJSONBody, url, normalizeData(data), structObject)
}

// Send ...
func (h *HTTP) Send(method, url string, data map[string]any) (res string, err error) {
	spec, err := h.buildRequestSpec(method, url, data)
	if err != nil {
		return "", err
	}
	return h.execute(spec)
}

func mapToString(data map[string]any) string {
	ret := make([]string, 0)
	for k, v := range data {
		ret = append(ret, k+"="+CompressStr(cast.ToString(v)))
	}
	return strings.Join(ret, "&")
}

// CompressStr 压缩字符串，去除空格或制表符
func CompressStr(str string) string {
	if str == "" {
		return ""
	}
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	return strings.Replace(str, "\r", "", -1)
}

type UploadFileStruct struct {
	Status int `json:"status"`
	Data   struct {
		Ext  string  `json:"ext"`
		Path string  `json:"path"`
		Size float64 `json:"size"`
		Src  string  `json:"src"`
	} `json:"data"`
	Message string `json:"message"`
}

func (h *HTTP) UploadFile(url string, rc io.Reader, fieldName string) (res UploadFileStruct, err error) {
	spec, err := h.buildUploadRequestSpec(url, rc, fieldName)
	if err != nil {
		return res, err
	}

	body, err := h.execute(spec)
	if err != nil {
		return res, err
	}

	err = jsonutil.JsonStringToAny(body, &res)
	return res, err
}

func doRequestWithTimeout(ctx context.Context, timeout time.Duration, req *fasthttp.Request, resp *fasthttp.Response) error {
	requestTimeout := timeout
	if requestTimeout <= 0 {
		requestTimeout = defaultTimeout
	}

	if ctx != nil {
		if deadline, ok := ctx.Deadline(); ok {
			deadlineTimeout := time.Until(deadline)
			if deadlineTimeout <= 0 {
				return ctx.Err()
			}
			if deadlineTimeout < requestTimeout {
				requestTimeout = deadlineTimeout
			}
		}
	}

	return fasthttp.DoTimeout(req, resp, requestTimeout)
}

func (h *HTTP) sendAndDecode(method, url string, data map[string]any, out any) (string, error) {
	res, err := h.Send(method, url, data)
	if err != nil || out == nil {
		return res, err
	}

	return res, jsonutil.JsonStringToAny(res, out)
}

func (h *HTTP) buildRequestSpec(method, url string, data map[string]any) (requestSpec, error) {
	args := mapToArgs(data)

	switch method {
	case sendMethodPost:
		return requestSpec{
			method:      httpMethodPost,
			url:         url,
			body:        []byte(args.String()),
			contentType: defaultContentType(method),
			logPayload:  "params:" + mapToString(data),
		}, nil
	case sendMethodJSONBody:
		body, err := json.Marshal(data)
		if err != nil {
			return requestSpec{}, err
		}
		return requestSpec{
			method:      httpMethodPost,
			url:         url,
			body:        body,
			contentType: defaultContentType(method),
			logPayload:  "jsonBody:" + string(body),
		}, nil
	case sendMethodDeleteBody:
		body, err := json.Marshal(data)
		if err != nil {
			return requestSpec{}, err
		}
		return requestSpec{
			method:      httpMethodDelete,
			url:         url,
			body:        body,
			contentType: defaultContentType(method),
			logPayload:  "jsonBody:" + string(body),
		}, nil
	default:
		return requestSpec{
			method:      httpMethodGet,
			url:         appendQuery(url, args.String()),
			contentType: defaultContentType(method),
			logPayload:  "params:" + mapToString(data),
		}, nil
	}
}

func (h *HTTP) buildUploadRequestSpec(url string, rc io.Reader, fieldName string) (requestSpec, error) {
	if err := h.contextErr(); err != nil {
		return requestSpec{}, err
	}
	if rc == nil {
		return requestSpec{}, errors.New("upload reader is nil")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fieldName)
	if err != nil {
		_ = writer.Close()
		return requestSpec{}, err
	}
	if _, err := io.Copy(part, rc); err != nil {
		_ = writer.Close()
		return requestSpec{}, err
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return requestSpec{}, err
	}

	return requestSpec{
		method:      httpMethodPost,
		url:         url,
		body:        body.Bytes(),
		contentType: contentType,
		logPayload:  "file:" + fieldName,
	}, nil
}

func (h *HTTP) execute(spec requestSpec) (string, error) {
	if err := h.contextErr(); err != nil {
		return "", err
	}

	startT := time.Now()
	entry := logs.WithContext(h.Ctx)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	h.prepareRequest(req, spec)
	logRequestStart(entry, spec)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := doRequestWithTimeout(h.Ctx, h.timeout, req, resp); err != nil {
		logRequestError(entry, spec, err)
		return "", err
	}

	body := string(resp.Body())
	if err := parseResponse(resp); err != nil {
		logRequestFailure(entry, spec, resp.StatusCode(), body)
		return "", err
	}

	logRequestSuccess(entry, spec, body, time.Since(startT))
	return body, nil
}

func (h *HTTP) prepareRequest(req *fasthttp.Request, spec requestSpec) {
	req.SetRequestURI(spec.url)
	req.Header.SetMethod(spec.method)
	if len(spec.body) > 0 {
		req.SetBody(spec.body)
	}

	applyRequestHeaders(req, h.header)
	addRequestIDHeader(req, h.Ctx)

	contentType := spec.contentType
	if h.contentType != "" {
		contentType = h.contentType
	}
	if contentType != "" {
		req.Header.SetContentType(contentType)
	}
}

func (h *HTTP) contextErr() error {
	if h.Ctx != nil && h.Ctx.Err() != nil {
		return h.Ctx.Err()
	}
	return nil
}

func parseResponse(resp *fasthttp.Response) error {
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("unexpected http status: %d", resp.StatusCode())
	}
	return nil
}

func mapToArgs(data map[string]any) *fasthttp.Args {
	args := &fasthttp.Args{}
	for k, v := range data {
		args.Add(k, cast.ToString(v))
	}
	return args
}

func normalizeData(data any) map[string]any {
	if data == nil {
		return nil
	}
	return cast.ToStringMap(data)
}

func logRequestStart(logger anyLogger, spec requestSpec) {
	logger.Info(spec.method + "--->|url:" + spec.url + " |" + spec.logPayload)
}

func logRequestError(logger anyLogger, spec requestSpec, err error) {
	logger.Error(spec.method + "--->|url:" + spec.url + " |" + spec.logPayload + "| error:" + err.Error())
}

func logRequestFailure(logger anyLogger, spec requestSpec, statusCode int, body string) {
	logger.Error("[" + cast.ToString(statusCode) + "]|" + spec.method + "--->|url:" + spec.url + " |" + spec.logPayload + " |res:" + CompressStr(body))
}

func logRequestSuccess(logger anyLogger, spec requestSpec, body string, elapsed time.Duration) {
	infoStr := "[" + cast.ToString(elapsed.Seconds()) + "]|" + spec.method + "--->|url:" + spec.url + " |" + spec.logPayload
	if os.Getenv("environment") != "prod" {
		infoStr += " |res:" + CompressStr(body)
	}
	logger.Info(infoStr)
}

type anyLogger interface {
	Info(...interface{})
	Error(...interface{})
}

func applyRequestHeaders(req *fasthttp.Request, headers map[string]string) {
	if headers == nil {
		return
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
}

func addRequestIDHeader(req *fasthttp.Request, ctx context.Context) {
	if ctx == nil {
		return
	}
	requestID := cast.ToString(ctx.Value("X-Request-Id"))
	if requestID != "" {
		req.Header.Add("X-Request-Id", requestID)
	}
}

func defaultContentType(method string) string {
	switch method {
	case "POST":
		return "application/x-www-form-urlencoded"
	case "JsonBody", "DELETEBody":
		return "application/json"
	default:
		return "application/json"
	}
}

func appendQuery(rawURL, query string) string {
	if query == "" {
		return rawURL
	}
	if strings.Contains(rawURL, "?") {
		if strings.HasSuffix(rawURL, "?") || strings.HasSuffix(rawURL, "&") {
			return rawURL + query
		}
		return rawURL + "&" + query
	}
	return rawURL + "?" + query
}
