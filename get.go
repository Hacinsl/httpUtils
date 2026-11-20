package httputils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Get 提交给定的查询参数，把返回体反序列化为给定的类型。
func Get[T any](u string, form *url.Values, cookies *map[string]string, headers *map[string]string) (*T, error) {
	resp, err := doGet(u, form, cookies, headers)
	if err != nil {
		return nil, err
	}

	return decode[T](resp, u)
}

// GetHtml 取回页面的 HTML 文本。
func GetHtml(u string, headers *map[string]string) (*string, error) {
	resp, err := doGet(u, nil, nil, headers)
	if err != nil {
		return nil, err
	}

	r, err := readBody(resp, u)
	if err != nil {
		return nil, err
	}

	html := string(r)
	return &html, nil
}

// GetLoaction 提交给定的查询参数，取回响应里的 Location。
func GetLoaction(u string, form *url.Values, cookies *map[string]string, headers *map[string]string) (*url.URL, error) {
	resp, err := doGet(u, form, cookies, headers)
	if err != nil {
		return nil, err
	}

	return resp.Location()
}

// GetCookie 提交给定的查询参数，取回响应里的 Set-Cookie。
func GetCookie(u string, form *url.Values, cookies *map[string]string, headers *map[string]string) ([]*http.Cookie, error) {
	resp, err := doGet(u, form, cookies, headers)
	if err != nil {
		return nil, err
	}

	return resp.Cookies(), nil
}

// GetResp 直接发送给定的请求体，返回一个 *http.Response。
// 返回的 Body 是一个可重复读的内存缓冲，无需 Close。
func GetResp(u string, form *url.Values, cookies *map[string]string, headers *map[string]string) (*http.Response, error) {
	return doGet(u, form, cookies, headers)
}

func doGet(u string, form *url.Values, cookies *map[string]string, headers *map[string]string) (*http.Response, error) {
	pu, _ := url.Parse(u)
	if form != nil {
		pu.RawQuery = form.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, pu.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request for %s: %w", pu.String(), err)
	}

	if cookies != nil {
		for k, v := range *cookies {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}

	if headers != nil {
		for k, v := range *headers {
			req.Header.Add(k, v)
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	dumpRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to get %s: %w", u, err)
	}

	// 把 body 换成与 TCP 连接无关的内存缓冲，使调用方仍能按 *http.Response 惯例读取
	r, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed to read response for %s: %w", u, err)
	}
	resp.Body = io.NopCloser(bytes.NewReader(r))

	dumpResponse(resp)

	if err := checkStatus(resp); err != nil {
		return resp, err // 即使状态码返回错误，依然返回response，便于提取错误信息
	}

	return resp, nil
}
