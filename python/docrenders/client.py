from __future__ import annotations

import os
from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional, Tuple
import urllib.request
import urllib.error
import json

DEFAULT_BASE_URL = "https://www.docrenders.com"


class DocRendersError(Exception):
    def __init__(self, code: str, message: str):
        super().__init__(message)
        self.code = code


@dataclass
class RenderOptions:
    format: Optional[str] = None          # "A4", "Letter", "Legal"
    margin_top: Optional[str] = None
    margin_right: Optional[str] = None
    margin_bottom: Optional[str] = None
    margin_left: Optional[str] = None
    landscape: bool = False

    def to_dict(self) -> dict:
        d = {}
        if self.format:
            d["format"] = self.format
        if self.margin_top:
            d["margin_top"] = self.margin_top
        if self.margin_right:
            d["margin_right"] = self.margin_right
        if self.margin_bottom:
            d["margin_bottom"] = self.margin_bottom
        if self.margin_left:
            d["margin_left"] = self.margin_left
        if self.landscape:
            d["landscape"] = True
        return d


@dataclass
class RenderRequest:
    markdown: Optional[str] = None
    html: Optional[str] = None
    template: Optional[str] = None
    options: RenderOptions = field(default_factory=RenderOptions)
    filename: Optional[str] = None


@dataclass
class RenderFileRequest:
    filename: str
    content: bytes
    options: RenderOptions = field(default_factory=RenderOptions)


@dataclass
class SignedURLResult:
    url: str
    expires_at: str
    render_time_ms: int


@dataclass
class RateLimit:
    requests_per_minute: int


@dataclass
class RenderUsage:
    used: int
    limit: int
    period: str


@dataclass
class UsageResult:
    key_prefix: str
    plan: str
    rate_limit: RateLimit
    renders: RenderUsage


class DocRendersClient:
    """Client for the DocRenders API."""

    def __init__(self, api_key: str, base_url: str = DEFAULT_BASE_URL):
        self.api_key = api_key
        self.base_url = base_url.rstrip("/")

    def _auth_header(self) -> dict:
        return {"Authorization": f"Bearer {self.api_key}"}

    def _post_json(self, path: str, payload: dict) -> urllib.request.Request:
        data = json.dumps(payload).encode()
        req = urllib.request.Request(
            self.base_url + path,
            data=data,
            headers={**self._auth_header(), "Content-Type": "application/json"},
            method="POST",
        )
        return req

    def _raise_for_response(self, e: urllib.error.HTTPError) -> None:
        try:
            body = json.loads(e.read())
            code = body.get("error", {}).get("code", "unknown_error")
            message = body.get("error", {}).get("message", str(e))
        except Exception:
            code, message = "unknown_error", str(e)
        raise DocRendersError(code, message)

    def _render_payload(self, req: RenderRequest, output: str) -> dict:
        payload: dict = {"output": output}
        if req.markdown:
            payload["markdown"] = req.markdown
        if req.html:
            payload["html"] = req.html
        if req.template:
            payload["template"] = req.template
        if req.filename:
            payload["filename"] = req.filename
        opts = req.options.to_dict()
        if opts:
            payload["options"] = opts
        return payload

    def render(self, req: RenderRequest) -> bytes:
        """Render Markdown or HTML to PDF. Returns raw PDF bytes."""
        http_req = self._post_json("/render", self._render_payload(req, "binary"))
        try:
            with urllib.request.urlopen(http_req) as resp:
                return resp.read()
        except urllib.error.HTTPError as e:
            self._raise_for_response(e)

    def render_signed_url(self, req: RenderRequest) -> SignedURLResult:
        """Render Markdown or HTML to PDF. Returns a signed download URL (expires in 15 min)."""
        http_req = self._post_json("/render", self._render_payload(req, "url"))
        try:
            with urllib.request.urlopen(http_req) as resp:
                body = json.loads(resp.read())
                return SignedURLResult(
                    url=body["url"],
                    expires_at=body["expires_at"],
                    render_time_ms=body["render_time_ms"],
                )
        except urllib.error.HTTPError as e:
            self._raise_for_response(e)

    def render_file(self, req: RenderFileRequest) -> bytes:
        """Upload a Markdown or HTML file and render it to PDF. Returns raw PDF bytes."""
        data, content_type = _build_multipart(req, "binary")
        http_req = urllib.request.Request(
            self.base_url + "/render/file",
            data=data,
            headers={**self._auth_header(), "Content-Type": content_type},
            method="POST",
        )
        try:
            with urllib.request.urlopen(http_req) as resp:
                return resp.read()
        except urllib.error.HTTPError as e:
            self._raise_for_response(e)

    def render_file_signed_url(self, req: RenderFileRequest) -> SignedURLResult:
        """Upload a Markdown or HTML file and render it to PDF. Returns a signed download URL."""
        data, content_type = _build_multipart(req, "url")
        http_req = urllib.request.Request(
            self.base_url + "/render/file",
            data=data,
            headers={**self._auth_header(), "Content-Type": content_type},
            method="POST",
        )
        try:
            with urllib.request.urlopen(http_req) as resp:
                body = json.loads(resp.read())
                return SignedURLResult(
                    url=body["url"],
                    expires_at=body["expires_at"],
                    render_time_ms=body["render_time_ms"],
                )
        except urllib.error.HTTPError as e:
            self._raise_for_response(e)

    def usage(self) -> UsageResult:
        """Return current period usage for the authenticated account."""
        http_req = urllib.request.Request(
            self.base_url + "/usage",
            headers=self._auth_header(),
            method="GET",
        )
        try:
            with urllib.request.urlopen(http_req) as resp:
                body = json.loads(resp.read())
                r = body["renders"]
                rl = body["rate_limit"]
                return UsageResult(
                    key_prefix=body["key_prefix"],
                    plan=body["plan"],
                    rate_limit=RateLimit(requests_per_minute=rl["requests_per_minute"]),
                    renders=RenderUsage(used=r["used"], limit=r["limit"], period=r["period"]),
                )
        except urllib.error.HTTPError as e:
            self._raise_for_response(e)


def _build_multipart(req: RenderFileRequest, output: str) -> Tuple[bytes, str]:
    boundary = b"----DocRendersBoundary7MA4YWxkTrZu0gW"
    lines = []

    def field(name: str, value: str) -> None:
        lines.append(b"--" + boundary)
        lines.append(f'Content-Disposition: form-data; name="{name}"'.encode())
        lines.append(b"")
        lines.append(value.encode())

    lines.append(b"--" + boundary)
    lines.append(
        f'Content-Disposition: form-data; name="file"; filename="{req.filename}"'.encode()
    )
    lines.append(b"Content-Type: application/octet-stream")
    lines.append(b"")
    lines.append(req.content)

    field("output", output)
    opts = req.options
    if opts.format:
        field("format", opts.format)
    if opts.margin_top:
        field("margin_top", opts.margin_top)
    if opts.margin_right:
        field("margin_right", opts.margin_right)
    if opts.margin_bottom:
        field("margin_bottom", opts.margin_bottom)
    if opts.margin_left:
        field("margin_left", opts.margin_left)
    if opts.landscape:
        field("landscape", "true")

    lines.append(b"--" + boundary + b"--")
    body = b"\r\n".join(lines)
    content_type = f"multipart/form-data; boundary={boundary.decode()}"
    return body, content_type
