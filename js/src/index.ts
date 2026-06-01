const DEFAULT_BASE_URL = "https://docrenders.com";

export interface RenderOptions {
  format?: "A4" | "Letter" | "Legal";
  margin_top?: string;
  margin_right?: string;
  margin_bottom?: string;
  margin_left?: string;
  landscape?: boolean;
}

export interface RenderRequest {
  markdown?: string;
  html?: string;
  options?: RenderOptions;
  template?: string;
}

export interface RenderFileRequest {
  file: Blob | File;
  filename?: string;
  options?: RenderOptions;
}

export interface SignedURLResult {
  url: string;
  expires_at: string;
  render_time_ms: number;
}

export interface UsageResult {
  key_prefix: string;
  plan: string;
  rate_limit: { requests_per_minute: number };
  renders: { used: number; limit: number; period: string };
}

export interface ClientOptions {
  baseURL?: string;
}

export class DocRendersError extends Error {
  constructor(public code: string, message: string) {
    super(message);
    this.name = "DocRendersError";
  }
}

export class DocRendersClient {
  private apiKey: string;
  private baseURL: string;

  constructor(apiKey: string, opts: ClientOptions = {}) {
    this.apiKey = apiKey;
    this.baseURL = opts.baseURL ?? DEFAULT_BASE_URL;
  }

  private get headers(): HeadersInit {
    return { Authorization: `Bearer ${this.apiKey}` };
  }

  private async handleError(res: Response): Promise<never> {
    const body = await res.json().catch(() => ({}));
    const code = body?.error?.code ?? "unknown_error";
    const message = body?.error?.message ?? `Unexpected status ${res.status}`;
    throw new DocRendersError(code, message);
  }

  /** Render Markdown or HTML to PDF. Returns raw PDF bytes. */
  async render(req: RenderRequest): Promise<Uint8Array> {
    const res = await fetch(`${this.baseURL}/render`, {
      method: "POST",
      headers: { ...this.headers, "Content-Type": "application/json" },
      body: JSON.stringify({ ...req, output: "binary" }),
    });
    if (!res.ok) await this.handleError(res);
    return new Uint8Array(await res.arrayBuffer());
  }

  /** Render Markdown or HTML to PDF. Returns a signed download URL (expires in 15 min). */
  async renderSignedURL(req: RenderRequest): Promise<SignedURLResult> {
    const res = await fetch(`${this.baseURL}/render`, {
      method: "POST",
      headers: { ...this.headers, "Content-Type": "application/json" },
      body: JSON.stringify({ ...req, output: "url" }),
    });
    if (!res.ok) await this.handleError(res);
    return res.json() as Promise<SignedURLResult>;
  }

  /** Upload a Markdown or HTML file and render it to PDF. Returns raw PDF bytes. */
  async renderFile(req: RenderFileRequest): Promise<Uint8Array> {
    const res = await fetch(`${this.baseURL}/render/file`, {
      method: "POST",
      headers: this.headers,
      body: buildFormData(req, "binary"),
    });
    if (!res.ok) await this.handleError(res);
    return new Uint8Array(await res.arrayBuffer());
  }

  /** Upload a Markdown or HTML file and render it to PDF. Returns a signed download URL. */
  async renderFileSignedURL(req: RenderFileRequest): Promise<SignedURLResult> {
    const res = await fetch(`${this.baseURL}/render/file`, {
      method: "POST",
      headers: this.headers,
      body: buildFormData(req, "url"),
    });
    if (!res.ok) await this.handleError(res);
    return res.json() as Promise<SignedURLResult>;
  }

  /** Return current period usage for the authenticated account. */
  async usage(): Promise<UsageResult> {
    const res = await fetch(`${this.baseURL}/usage`, {
      headers: this.headers,
    });
    if (!res.ok) await this.handleError(res);
    return res.json() as Promise<UsageResult>;
  }
}

function buildFormData(req: RenderFileRequest, output: string): FormData {
  const fd = new FormData();
  const filename = req.filename ?? (req.file instanceof File ? req.file.name : "upload.md");
  fd.append("file", req.file, filename);
  fd.append("output", output);
  if (req.options?.format) fd.append("format", req.options.format);
  if (req.options?.margin_top) fd.append("margin_top", req.options.margin_top);
  if (req.options?.margin_right) fd.append("margin_right", req.options.margin_right);
  if (req.options?.margin_bottom) fd.append("margin_bottom", req.options.margin_bottom);
  if (req.options?.margin_left) fd.append("margin_left", req.options.margin_left);
  if (req.options?.landscape) fd.append("landscape", "true");
  return fd;
}

export default DocRendersClient;
