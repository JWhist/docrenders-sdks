package pdfgen_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pdfgen "github.com/JWhist/pdfgen-sdks/go"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *pdfgen.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := pdfgen.NewClient("pdg_test_key", pdfgen.WithBaseURL(srv.URL))
	return srv, client
}

func TestRender_Binary(t *testing.T) {
	want := []byte("%PDF-1.4 fake pdf bytes")
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/render" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer pdg_test_key" {
			t.Errorf("missing or wrong auth header")
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["output"] != "binary" {
			t.Errorf("expected output=binary, got %v", body["output"])
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(want)
	})

	got, err := client.Render(context.Background(), pdfgen.RenderRequest{
		Markdown: "# Hello",
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("body mismatch: got %q, want %q", got, want)
	}
}

func TestRenderSignedURL(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["output"] != "url" {
			t.Errorf("expected output=url, got %v", body["output"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":            "https://storage.example.com/pdfs/user/abc.pdf?sig=xyz",
			"expires_at":     "2026-05-31T15:00:00Z",
			"render_time_ms": 1200,
		})
	})

	result, err := client.RenderSignedURL(context.Background(), pdfgen.RenderRequest{
		Markdown: "# Hello",
	})
	if err != nil {
		t.Fatalf("RenderSignedURL: %v", err)
	}
	if result.URL == "" {
		t.Error("expected non-empty URL")
	}
	if result.RenderTimeMs != 1200 {
		t.Errorf("render_time_ms: got %d, want 1200", result.RenderTimeMs)
	}
}

func TestRenderFile_Binary(t *testing.T) {
	want := []byte("%PDF-1.4 fake pdf bytes")
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/render/file" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseMultipartForm(1 << 20)
		if r.FormValue("output") != "binary" {
			t.Errorf("expected output=binary")
		}
		if _, _, err := r.FormFile("file"); err != nil {
			t.Errorf("missing file field: %v", err)
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(want)
	})

	got, err := client.RenderFile(context.Background(), pdfgen.RenderFileRequest{
		Filename: "invoice.md",
		Content:  []byte("# Invoice"),
	})
	if err != nil {
		t.Fatalf("RenderFile: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("body mismatch: got %q, want %q", got, want)
	}
}

func TestRenderFileSignedURL(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1 << 20)
		if r.FormValue("output") != "url" {
			t.Errorf("expected output=url")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":            "https://storage.example.com/pdfs/user/abc.pdf",
			"expires_at":     "2026-05-31T15:00:00Z",
			"render_time_ms": 900,
		})
	})

	result, err := client.RenderFileSignedURL(context.Background(), pdfgen.RenderFileRequest{
		Filename: "report.md",
		Content:  []byte("# Report"),
	})
	if err != nil {
		t.Fatalf("RenderFileSignedURL: %v", err)
	}
	if result.URL == "" {
		t.Error("expected non-empty URL")
	}
}

func TestUsage(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/usage" || r.Method != http.MethodGet {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"plan":              "starter",
			"period_start":      "2026-05-01T00:00:00Z",
			"period_end":        "2026-06-01T00:00:00Z",
			"renders_used":      42,
			"renders_limit":     5000,
			"renders_remaining": 4958,
		})
	})

	usage, err := client.Usage(context.Background())
	if err != nil {
		t.Fatalf("Usage: %v", err)
	}
	if usage.Plan != "starter" {
		t.Errorf("plan: got %q, want %q", usage.Plan, "starter")
	}
	if usage.RendersUsed != 42 {
		t.Errorf("renders_used: got %d, want 42", usage.RendersUsed)
	}
}

func TestAPIError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "quota_exceeded",
				"message": "monthly render limit of 100 reached",
			},
		})
	})

	_, err := client.Render(context.Background(), pdfgen.RenderRequest{Markdown: "# Hello"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if want := "pdfgen: quota_exceeded: monthly render limit of 100 reached"; err.Error() != want {
		t.Errorf("error message: got %q, want %q", err.Error(), want)
	}
}
