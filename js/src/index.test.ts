/**
 * Tests for the DocRenders JS SDK.
 * Uses a mock fetch to avoid real network calls.
 */

import DocRendersClient, { DocRendersError } from "./index";

function mockFetch(responses: Array<{ status: number; body: unknown; contentType?: string }>) {
  let call = 0;
  return async (url: string, init?: RequestInit): Promise<Response> => {
    const r = responses[call++];
    const isJson = !r.contentType || r.contentType === "application/json";
    const bodyBytes =
      r.body instanceof Uint8Array
        ? Buffer.from(r.body)
        : Buffer.from(JSON.stringify(r.body));

    return new Response(bodyBytes as unknown as BodyInit, {
      status: r.status,
      headers: { "Content-Type": r.contentType ?? "application/json" },
    });
  };
}

const PDF_BYTES = new Uint8Array([0x25, 0x50, 0x44, 0x46]); // %PDF

describe("DocRendersClient", () => {
  test("render() sends output=binary and returns bytes", async () => {
    const client = new DocRendersClient("dcr_test_key");
    (global as any).fetch = mockFetch([
      { status: 200, body: PDF_BYTES, contentType: "application/pdf" },
    ]);

    const result = await client.render({ markdown: "# Hello" });
    expect(result).toBeInstanceOf(Uint8Array);
    expect(result[0]).toBe(0x25); // %
  });

  test("renderSignedURL() sends output=url and returns url/expires_at", async () => {
    const client = new DocRendersClient("dcr_test_key");
    (global as any).fetch = mockFetch([
      {
        status: 200,
        body: {
          url: "https://storage.example.com/abc.pdf",
          expires_at: "2026-05-31T15:00:00Z",
          render_time_ms: 1200,
        },
      },
    ]);

    const result = await client.renderSignedURL({ markdown: "# Hello" });
    expect(result.url).toBe("https://storage.example.com/abc.pdf");
    expect(result.render_time_ms).toBe(1200);
  });

  test("renderFile() sends multipart with output=binary", async () => {
    const client = new DocRendersClient("dcr_test_key");
    (global as any).fetch = mockFetch([
      { status: 200, body: PDF_BYTES, contentType: "application/pdf" },
    ]);

    const result = await client.renderFile({
      file: new Blob(["# Invoice"], { type: "text/markdown" }),
      filename: "invoice.md",
    });
    expect(result).toBeInstanceOf(Uint8Array);
  });

  test("renderFileSignedURL() returns signed URL", async () => {
    const client = new DocRendersClient("dcr_test_key");
    (global as any).fetch = mockFetch([
      {
        status: 200,
        body: {
          url: "https://storage.example.com/xyz.pdf",
          expires_at: "2026-05-31T15:00:00Z",
          render_time_ms: 900,
        },
      },
    ]);

    const result = await client.renderFileSignedURL({
      file: new Blob(["# Report"]),
      filename: "report.md",
    });
    expect(result.url).toBe("https://storage.example.com/xyz.pdf");
  });

  test("usage() returns usage data", async () => {
    const client = new DocRendersClient("dcr_test_key");
    (global as any).fetch = mockFetch([
      {
        status: 200,
        body: {
          plan: "starter",
          period_start: "2026-05-01T00:00:00Z",
          period_end: "2026-06-01T00:00:00Z",
          renders_used: 42,
          renders_limit: 5000,
          renders_remaining: 4958,
        },
      },
    ]);

    const result = await client.usage();
    expect(result.plan).toBe("starter");
    expect(result.renders_used).toBe(42);
  });

  test("throws DocRendersError on API error", async () => {
    const client = new DocRendersClient("dcr_test_key");
    (global as any).fetch = mockFetch([
      {
        status: 429,
        body: { error: { code: "quota_exceeded", message: "monthly render limit reached" } },
      },
    ]);

    let caught: unknown;
    try {
      await client.render({ markdown: "# Hello" });
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(DocRendersError);
    expect((caught as DocRendersError).code).toBe("quota_exceeded");
    expect((caught as DocRendersError).message).toBe("monthly render limit reached");
  });

  test("sets Authorization header on every request", async () => {
    let capturedHeaders: Headers | undefined;
    (global as any).fetch = async (_url: string, init?: RequestInit) => {
      capturedHeaders = new Headers(init?.headers);
      return new Response(PDF_BYTES, {
        status: 200,
        headers: { "Content-Type": "application/pdf" },
      });
    };

    const client = new DocRendersClient("dcr_test_key");
    await client.render({ markdown: "# Hello" });
    expect(capturedHeaders?.get("Authorization")).toBe("Bearer dcr_test_key");
  });
});
