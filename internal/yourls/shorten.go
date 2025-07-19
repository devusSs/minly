package yourls

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (c *Client) Shorten(ctx context.Context, original string) (string, error) {
	if ctx == nil {
		return "", errors.New("context cannot be nil")
	}

	if original == "" {
		return "", errors.New("original cannot be empty")
	}

	keyword, err := generateKeyword()
	if err != nil {
		return "", fmt.Errorf("failed to generate keyword: %w", err)
	}

	v := url.Values{}
	v.Set("signature", c.signature)
	v.Set("action", shortenAction)
	v.Set("url", original)
	v.Set("format", shortenFormat)
	v.Set("title", c.title)
	v.Set("keyword", keyword)

	var req *http.Request
	req, err = http.NewRequestWithContext(
		ctx,
		shortenHTTPMethod,
		c.endpoint,
		strings.NewReader(v.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var resp *http.Response
	resp, err = c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var res shortenResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if res.Status != "success" {
		return "", fmt.Errorf(
			"shorten error: %s (code: %s, message: %s, errorCode: %s, statusCode: %s)",
			res.Status,
			res.Code,
			res.Message,
			res.ErrorCode,
			res.StatusCode,
		)
	}

	return res.Shorturl, nil
}

const (
	shortenHTTPMethod = http.MethodPost
	shortenAction     = "shorturl"
	shortenFormat     = "json"
)

type shortenResponse struct {
	Status     string `json:"status"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	ErrorCode  string `json:"errorCode"`
	StatusCode string `json:"statusCode"`
	Title      string `json:"title"`
	Shorturl   string `json:"shorturl"`
}

func generateKeyword() (string, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}

	return uid.String(), nil
}
