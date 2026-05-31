// Package pdfgen is the official Go SDK for the DocRenders API.
// It provides methods to render Markdown or HTML to PDF and retrieve usage information.
//
// Usage:
//
//	client := docrenders.NewClient("dcr_live_YOUR_API_KEY")
//	pdf, err := client.Render(ctx, docrenders.RenderRequest{Markdown: "# Hello"})
package docrenders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

const defaultBaseURL = "https://docrenders.com"

// Client is the DocRenders API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new DocRenders client with the given API key.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Option configures the client.
type Option func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient overrides the underlying HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// RenderOptions controls PDF output settings.
type RenderOptions struct {
	Format        string `json:"format,omitempty"`
	MarginTop     string `json:"margin_top,omitempty"`
	MarginRight   string `json:"margin_right,omitempty"`
	MarginBottom  string `json:"margin_bottom,omitempty"`
	MarginLeft    string `json:"margin_left,omitempty"`
	Landscape     bool   `json:"landscape,omitempty"`
}

// RenderRequest is the input for Render and RenderSignedURL.
type RenderRequest struct {
	Markdown string        `json:"markdown,omitempty"`
	HTML     string        `json:"html,omitempty"`
	Options  RenderOptions `json:"options,omitempty"`
	Template string        `json:"template,omitempty"`
}

// RenderFileRequest is the input for RenderFile and RenderFileSignedURL.
type RenderFileRequest struct {
	// Filename is used to determine the content type (.md / .html).
	Filename string
	// Content is the file contents.
	Content  []byte
	Options  RenderOptions
}

// SignedURLResult is returned by RenderSignedURL and RenderFileSignedURL.
type SignedURLResult struct {
	URL          string    `json:"url"`
	ExpiresAt    time.Time `json:"expires_at"`
	RenderTimeMs int64     `json:"render_time_ms"`
}

// UsageResult is returned by Usage.
type UsageResult struct {
	Plan              string    `json:"plan"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	RendersUsed       int       `json:"renders_used"`
	RendersLimit      int       `json:"renders_limit"`
	RendersRemaining  int       `json:"renders_remaining"`
}

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var apiErr apiError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Error.Code != "" {
			return nil, fmt.Errorf("docrenders: %s: %s", apiErr.Error.Code, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("docrenders: unexpected status %d", resp.StatusCode)
	}
	return resp, nil
}

// Render converts Markdown or HTML to a PDF and returns the raw bytes.
func (c *Client) Render(ctx context.Context, req RenderRequest) ([]byte, error) {
	body := renderBody(req, "binary")
	httpReq, err := c.newJSONRequest(ctx, http.MethodPost, "/render", body)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// RenderSignedURL converts Markdown or HTML to a PDF, stores it, and returns a signed download URL.
func (c *Client) RenderSignedURL(ctx context.Context, req RenderRequest) (*SignedURLResult, error) {
	body := renderBody(req, "url")
	httpReq, err := c.newJSONRequest(ctx, http.MethodPost, "/render", body)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result SignedURLResult
	return &result, json.NewDecoder(resp.Body).Decode(&result)
}

// RenderFile uploads a Markdown or HTML file and returns the rendered PDF as raw bytes.
func (c *Client) RenderFile(ctx context.Context, req RenderFileRequest) ([]byte, error) {
	body, ct, err := renderFileBody(req, "binary")
	if err != nil {
		return nil, err
	}
	httpReq, err := c.newMultipartRequest(ctx, body, ct)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// RenderFileSignedURL uploads a Markdown or HTML file and returns a signed download URL.
func (c *Client) RenderFileSignedURL(ctx context.Context, req RenderFileRequest) (*SignedURLResult, error) {
	body, ct, err := renderFileBody(req, "url")
	if err != nil {
		return nil, err
	}
	httpReq, err := c.newMultipartRequest(ctx, body, ct)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result SignedURLResult
	return &result, json.NewDecoder(resp.Body).Decode(&result)
}

// Usage returns the current period usage for the authenticated account.
func (c *Client) Usage(ctx context.Context) (*UsageResult, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/usage", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result UsageResult
	return &result, json.NewDecoder(resp.Body).Decode(&result)
}

func (c *Client) newJSONRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) newMultipartRequest(ctx context.Context, body *bytes.Buffer, contentType string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/render/file", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

func renderBody(req RenderRequest, output string) map[string]interface{} {
	body := map[string]interface{}{"output": output}
	if req.Markdown != "" {
		body["markdown"] = req.Markdown
	}
	if req.HTML != "" {
		body["html"] = req.HTML
	}
	if req.Template != "" {
		body["template"] = req.Template
	}
	if (req.Options != RenderOptions{}) {
		body["options"] = req.Options
	}
	return body
}

func renderFileBody(req RenderFileRequest, output string) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filepath.Base(req.Filename))
	if err != nil {
		return nil, "", err
	}
	if _, err := fw.Write(req.Content); err != nil {
		return nil, "", err
	}
	if req.Options.Format != "" {
		w.WriteField("format", req.Options.Format)
	}
	if req.Options.MarginTop != "" {
		w.WriteField("margin_top", req.Options.MarginTop)
	}
	if req.Options.MarginRight != "" {
		w.WriteField("margin_right", req.Options.MarginRight)
	}
	if req.Options.MarginBottom != "" {
		w.WriteField("margin_bottom", req.Options.MarginBottom)
	}
	if req.Options.MarginLeft != "" {
		w.WriteField("margin_left", req.Options.MarginLeft)
	}
	if req.Options.Landscape {
		w.WriteField("landscape", "true")
	}
	w.WriteField("output", output)
	w.Close()
	return &buf, w.FormDataContentType(), nil
}
