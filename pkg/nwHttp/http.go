package nwHttp

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/valyala/fasthttp"

	"github.com/silen/nuwa/pkg/logs"
	"github.com/silen/nuwa/pkg/tools"
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
	contentType []string
	Ctx         context.Context
}

// Get ..
func (h *HTTP) SetHeader(data map[string]string) *HTTP {
	h.header = data
	return h
}

// Get ..
func (h *HTTP) SetContentType(data string) *HTTP {
	if h.contentType == nil {
		h.contentType = make([]string, 0)
	}
	h.contentType = append(h.contentType, data)
	return h
}

// Get ..
func (h *HTTP) Get(url string, data any, structObject any) (res string, err error) {

	res, err = h.Send("GET", url, cast.ToStringMap(data))
	if err != nil {
		return
	}
	if structObject != nil {
		err = tools.JsonStringToAny(res, structObject)
	}

	return
}

// Post ..
func (h *HTTP) Post(url string, data any, structObject any) (res string, err error) {

	res, err = h.Send("POST", url, cast.ToStringMap(data))
	if err != nil {
		return
	}
	if structObject != nil {
		err = tools.JsonStringToAny(res, structObject)
	}

	return
}

// Post ..
func (h *HTTP) PostJson(url string, data any, structObject any) (res string, err error) {

	res, err = h.Send("JsonBody", url, cast.ToStringMap(data))
	if err != nil {
		return
	}

	if structObject != nil {
		err = tools.JsonStringToAny(res, structObject)
	}
	return
}

// Send ...
func (h *HTTP) Send(method, url string, data map[string]any) (res string, err error) {
	startT := time.Now()
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req) // 用完需要释放资源

	if h.header != nil {
		for k, v := range h.header {
			req.Header.Add(k, v)
		}
	}

	if h.contentType != nil {
		for _, v := range h.contentType {
			req.Header.SetContentType(v)
		}
	} else {
		req.Header.SetContentType("application/json")
	}

	//增加线程追踪
	if h.Ctx != nil && cast.ToString(h.Ctx.Value("X-Request-Id")) != "" {
		req.Header.Add("X-Request-Id", cast.ToString(h.Ctx.Value("X-Request-Id")))
	}

	var requestParams string
	args := &fasthttp.Args{}
	for k, v := range data {
		args.Add(k, cast.ToString(v))
	}
	switch method {
	case "POST":
		req.Header.SetMethod("POST")
		req.SetBodyString(args.String())
		requestParams = `params:` + mapToString(data)
	case "JsonBody":
		bytes, _ := json.Marshal(data)
		req.Header.SetMethod("POST")
		req.SetBody(bytes)
		requestParams = `jsonBody:` + string(bytes)
	case "DELETEBody":
		req.Header.SetMethod("DELETE")
		bytes, _ := json.Marshal(data)
		req.SetBody(bytes)
		requestParams = `jsonBody:` + string(bytes)
	default:
		req.Header.SetMethod("GET")
		if !strings.HasSuffix(url, "?") {
			url += "?"
		}

		url += args.String()
		requestParams = `params:` + mapToString(data)
	}

	infoStr := method + "--->|url:" + url + " |" + requestParams
	logs := logs.WithContext(h.Ctx)
	logs.Info(infoStr)

	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err = fasthttp.Do(req, resp); err != nil {
		infoStr := method + "--->|url:" + url + " |" + requestParams + "| error:" + err.Error()
		logs.Error(infoStr)
		return
	}

	res = string(resp.Body())
	curlTime := time.Since(startT).Seconds()

	//生产环境不打印
	if os.Getenv("environment") != "prod" {
		infoStr = "[" + cast.ToString(curlTime) + "]|" + method + "--->|url:" + url + " |" + requestParams + " |res:" + CompressStr(res)
	} else {
		infoStr = "[" + cast.ToString(curlTime) + "]|" + method + "--->|url:" + url + " |" + requestParams
	}
	logs.Info(infoStr)
	return
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
