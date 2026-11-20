package httputils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// readBody 读取响应体，并统一拼装读取失败的错误信息。
func readBody(resp *http.Response, u string) ([]byte, error) {
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response for %s: %w", u, err)
	}

	return r, nil
}

// decode 把响应体反序列化为给定的类型。
func decode[T any](resp *http.Response, u string) (*T, error) {
	r, err := readBody(resp, u)
	if err != nil {
		return nil, err
	}

	result := new(T)
	if err := json.Unmarshal(r, result); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal response for %s: %w", u, err)
	}
	return result, nil
}
