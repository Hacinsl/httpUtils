package httputils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func dumpRequest(req *http.Request) {
	if !EnableDebugOutput {
		return
	}

	fmt.Println()
	fmt.Printf("[%s] %s\n", req.Method, req.URL.String())

	fmt.Println("---Request Headers---")
	if req.Header != nil {
		for k, h := range req.Header {
			for _, v := range h {
				fmt.Printf("%s: %s\n", k, v)
			}
		}
	}

	fmt.Println("---Request Cookies---")
	if req.Cookies() != nil {
		for _, v := range req.Cookies() {
			fmt.Printf("%s: %s\n", v.Name, v.Value)
		}
	}

	fmt.Println("---Request Body---")
	if req.GetBody != nil {
		if rc, err := req.GetBody(); err == nil {
			body, _ := io.ReadAll(rc)
			rc.Close()
			fmt.Printf("%s\n", body)
		} else {
			fmt.Printf("(Failed to read request body: %v)\n", err)
		}
	} else if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(body))
		fmt.Printf("%s\n", body)
	} else {
		fmt.Println("(empty)")
	}
}

func dumpResponse(resp *http.Response) {
	if !EnableDebugOutput {
		return
	}

	fmt.Println()
	// 手工构造的 Response 可能没有 Request，此时只打印状态行。
	if resp.Request != nil {
		fmt.Printf("[%s] %s %s\n", resp.Request.Method, resp.Status, resp.Request.URL.String())
	} else {
		fmt.Printf("[%s]\n", resp.Status)
	}

	fmt.Println("---Response Headers---")
	if resp.Header != nil {
		for k, h := range resp.Header {
			for _, v := range h {
				fmt.Printf("%s: %s\n", k, v)
			}
		}
	}

	fmt.Println("---Response Cookies---")
	if resp.Cookies() != nil {
		for _, v := range resp.Cookies() {
			fmt.Printf("%s: %s\n", v.Name, v.Value)
		}
	}

	fmt.Println("---Response Body---")

	if resp.Body == nil {
		fmt.Println("(empty)")
		return
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(body))
	fmt.Printf("%s\n", body)
}
