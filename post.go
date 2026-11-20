package httputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// PostJSON 把 payload 序列化为 JSON 后提交，把返回体反序列化为给定的类型。
func PostJSON[T any](u string, payload interface{}, cookies *map[string]string, headers *map[string]string) (*T, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal json: %w", err)
	}

	resp, err := doPost(u, data, cookies, headers)
	if err != nil {
		return nil, err
	}

	return decode[T](resp, u)
}

// PostForm 把 form 编码为表单后提交，把返回体反序列化为给定的类型。
func PostForm[T any](u string, form *url.Values, cookies *map[string]string, headers *map[string]string) (*T, error) {
	resp, err := doPost(u, []byte(form.Encode()), cookies, headers)
	if err != nil {
		return nil, err
	}

	return decode[T](resp, u)
}

// PostCookie 把 form 编码为表单后提交，取回响应里的 Set-Cookie。
func PostCookie(u string, form url.Values, cookies *map[string]string, headers *map[string]string) ([]*http.Cookie, error) {
	resp, err := doPost(u, []byte(form.Encode()), cookies, headers)
	if err != nil {
		return nil, err
	}

	return resp.Cookies(), nil
}

// PostBytes 直接发送给定的请求体，返回返回体。
func PostBytes(u string, body []byte, cookies *map[string]string, headers *map[string]string) (*[]byte, error) {
	resp, err := doPost(u, body, cookies, headers)
	if err != nil {
		return nil, err
	}

	r, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response for %s: %w", u, err)
	}

	return &r, nil
}

// PostResp 直接发送给定的请求体，返回一个 *http.Response。
// 返回的 Body 是一个可重复读的内存缓冲，无需 Close。
func PostResp(u string, body []byte, cookies *map[string]string, headers *map[string]string) (*http.Response, error) {
	return doPost(u, body, cookies, headers)
}

func doPost(u string, body []byte, cookies *map[string]string, headers *map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Failed to create request for %s: %w", u, err)
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
		return nil, fmt.Errorf("Failed to post %s: %w", u, err)
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
