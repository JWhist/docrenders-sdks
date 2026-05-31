"""Tests for the DocRenders Python SDK using a mock HTTP server."""

import json
import threading
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer

from docrenders import DocRendersClient, DocRendersError, RenderOptions, RenderRequest, RenderFileRequest


def start_mock_server(handler_class):
    server = HTTPServer(("127.0.0.1", 0), handler_class)
    thread = threading.Thread(target=server.serve_forever)
    thread.daemon = True
    thread.start()
    return server


class TestRenderBinary(unittest.TestCase):
    def test_render_returns_bytes(self):
        class Handler(BaseHTTPRequestHandler):
            def do_POST(self):
                length = int(self.headers.get("Content-Length", 0))
                body = json.loads(self.rfile.read(length))
                assert body["output"] == "binary"
                assert self.headers.get("Authorization") == "Bearer dcr_test_key"
                resp = b"%PDF-1.4 fake"
                self.send_response(200)
                self.send_header("Content-Type", "application/pdf")
                self.send_header("Content-Length", str(len(resp)))
                self.end_headers()
                self.wfile.write(resp)

            def log_message(self, *args): pass

        server = start_mock_server(Handler)
        client = DocRendersClient("dcr_test_key", base_url=f"http://127.0.0.1:{server.server_address[1]}")
        result = client.render(RenderRequest(markdown="# Hello"))
        self.assertEqual(result, b"%PDF-1.4 fake")
        server.shutdown()


class TestRenderSignedURL(unittest.TestCase):
    def test_render_signed_url(self):
        class Handler(BaseHTTPRequestHandler):
            def do_POST(self):
                length = int(self.headers.get("Content-Length", 0))
                body = json.loads(self.rfile.read(length))
                assert body["output"] == "url"
                resp = json.dumps({
                    "url": "https://storage.example.com/abc.pdf",
                    "expires_at": "2026-05-31T15:00:00Z",
                    "render_time_ms": 1200,
                }).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(resp)))
                self.end_headers()
                self.wfile.write(resp)

            def log_message(self, *args): pass

        server = start_mock_server(Handler)
        client = DocRendersClient("dcr_test_key", base_url=f"http://127.0.0.1:{server.server_address[1]}")
        result = client.render_signed_url(RenderRequest(markdown="# Hello"))
        self.assertEqual(result.url, "https://storage.example.com/abc.pdf")
        self.assertEqual(result.render_time_ms, 1200)
        server.shutdown()


class TestRenderFile(unittest.TestCase):
    def test_render_file_binary(self):
        class Handler(BaseHTTPRequestHandler):
            def do_POST(self):
                # Read the multipart body — just verify it arrives and output=binary
                length = int(self.headers.get("Content-Length", 0))
                raw = self.rfile.read(length).decode(errors="replace")
                assert "output" in raw
                assert "binary" in raw
                assert "invoice.md" in raw
                resp = b"%PDF-1.4 fake"
                self.send_response(200)
                self.send_header("Content-Type", "application/pdf")
                self.send_header("Content-Length", str(len(resp)))
                self.end_headers()
                self.wfile.write(resp)

            def log_message(self, *args): pass

        server = start_mock_server(Handler)
        client = DocRendersClient("dcr_test_key", base_url=f"http://127.0.0.1:{server.server_address[1]}")
        result = client.render_file(RenderFileRequest(filename="invoice.md", content=b"# Invoice"))
        self.assertEqual(result, b"%PDF-1.4 fake")
        server.shutdown()


class TestUsage(unittest.TestCase):
    def test_usage(self):
        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                assert self.path == "/usage"
                resp = json.dumps({
                    "plan": "starter",
                    "period_start": "2026-05-01T00:00:00Z",
                    "period_end": "2026-06-01T00:00:00Z",
                    "renders_used": 42,
                    "renders_limit": 5000,
                    "renders_remaining": 4958,
                }).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(resp)))
                self.end_headers()
                self.wfile.write(resp)

            def log_message(self, *args): pass

        server = start_mock_server(Handler)
        client = DocRendersClient("dcr_test_key", base_url=f"http://127.0.0.1:{server.server_address[1]}")
        usage = client.usage()
        self.assertEqual(usage.plan, "starter")
        self.assertEqual(usage.renders_used, 42)
        self.assertEqual(usage.renders_remaining, 4958)
        server.shutdown()


class TestAPIError(unittest.TestCase):
    def test_api_error_raises(self):
        class Handler(BaseHTTPRequestHandler):
            def do_POST(self):
                resp = json.dumps({
                    "error": {"code": "quota_exceeded", "message": "monthly render limit reached"}
                }).encode()
                self.send_response(429)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(resp)))
                self.end_headers()
                self.wfile.write(resp)

            def log_message(self, *args): pass

        server = start_mock_server(Handler)
        client = DocRendersClient("dcr_test_key", base_url=f"http://127.0.0.1:{server.server_address[1]}")
        with self.assertRaises(DocRendersError) as cm:
            client.render(RenderRequest(markdown="# Hello"))
        self.assertEqual(cm.exception.code, "quota_exceeded")
        server.shutdown()


if __name__ == "__main__":
    unittest.main()
