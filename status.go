package httputils

import (
	"fmt"
	"net/http"
)

// checkStatus 校验响应的状态码，2xx 与 3xx（未跟随的重定向）视为成功。
func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}

	if resp.Request != nil && resp.Request.URL != nil {
		return fmt.Errorf("Unexpected status code: %d for %s", resp.StatusCode, resp.Request.URL.String())
	}
	return fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
}
