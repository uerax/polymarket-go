package clob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) get(endpoint string, options *RequestOptions) (any, error) {
	return c.request(http.MethodGet, endpoint, options, false)
}

func (c *Client) post(endpoint string, options *RequestOptions) (any, error) {
	return c.request(http.MethodPost, endpoint, options, c.retryOnError)
}

func (c *Client) put(endpoint string, options *RequestOptions) (any, error) {
	return c.request(http.MethodPut, endpoint, options, false)
}

func (c *Client) del(endpoint string, options *RequestOptions) (any, error) {
	return c.request(http.MethodDelete, endpoint, options, false)
}

func (c *Client) request(method string, endpoint string, options *RequestOptions, retry bool) (any, error) {
	res, status, err := c.doRequest(method, endpoint, options)
	if err != nil {
		if retry && method == http.MethodPost {
			time.Sleep(30 * time.Millisecond)
			res, status, err = c.doRequest(method, endpoint, options)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	if retry && method == http.MethodPost && isTransientRetryResult(res, status) {
		time.Sleep(30 * time.Millisecond)
		res, status, err = c.doRequest(method, endpoint, options)
		if err != nil {
			return nil, err
		}
	}

	if c.throwOnError {
		if m, ok := res.(map[string]any); ok {
			if _, hasErr := m["error"]; hasErr {
				msg := stringifyErrorValue(m["error"])
				return nil, &ApiError{Message: msg, Status: status, Data: res}
			}
		}
	}

	return res, nil
}

func (c *Client) doRequest(method string, endpoint string, options *RequestOptions) (any, int, error) {
	fullURL, err := addQueryParams(endpoint, options)
	if err != nil {
		return nil, 0, err
	}

	var bodyReader io.Reader
	if options != nil && options.Data != nil {
		switch v := options.Data.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		default:
			b, mErr := json.Marshal(v)
			if mErr != nil {
				return nil, 0, mErr
			}
			bodyReader = bytes.NewReader(b)
		}
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	applyHeaders(req, method, options)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return map[string]any{"error": err.Error()}, 0, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]any{"error": err.Error()}, resp.StatusCode, nil
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return parseErrorBody(body, resp.StatusCode), resp.StatusCode, nil
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return map[string]any{}, resp.StatusCode, nil
	}

	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, resp.StatusCode, err
	}
	return decoded, resp.StatusCode, nil
}

func addQueryParams(endpoint string, options *RequestOptions) (string, error) {
	if options == nil || len(options.Params) == 0 {
		return endpoint, nil
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range options.Params {
		if v == nil {
			continue
		}
		switch vv := v.(type) {
		case string:
			if vv != "" {
				q.Set(k, vv)
			}
		case []string:
			if len(vv) > 0 {
				q.Set(k, strings.Join(vv, ","))
			}
		default:
			q.Set(k, fmt.Sprintf("%v", vv))
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func applyHeaders(req *http.Request, method string, options *RequestOptions) {
	req.Header.Set("User-Agent", "@polymarket/clob-client")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	if method == http.MethodGet {
		req.Header.Set("Accept-Encoding", "gzip")
	}
	if options == nil || options.Headers == nil {
		return
	}
	for k, v := range options.Headers {
		req.Header.Set(k, v)
	}
}

func parseErrorBody(body []byte, status int) map[string]any {
	if len(bytes.TrimSpace(body)) == 0 {
		return map[string]any{"error": "", "status": status}
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return map[string]any{"error": string(body), "status": status}
	}
	if m, ok := decoded.(map[string]any); ok {
		if _, has := m["error"]; has {
			m["status"] = status
			return m
		}
		return map[string]any{"error": m, "status": status}
	}
	return map[string]any{"error": decoded, "status": status}
}

func stringifyErrorValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func isTransientRetryResult(res any, status int) bool {
	if status >= http.StatusInternalServerError && status < 600 {
		return true
	}
	if status != 0 {
		return false
	}
	m, ok := res.(map[string]any)
	if !ok {
		return false
	}
	_, hasError := m["error"]
	return hasError
}
