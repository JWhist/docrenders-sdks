# docrenders-sdk (JavaScript / TypeScript)

Official JS/TS SDK for the [DocRenders API](https://docrenders.com).

## Installation

```bash
npm install docrenders-sdk
```

## Usage

```ts
import DocRendersClient from "docrenders-sdk";

const client = new DocRendersClient(process.env.PDFGEN_API_KEY!);

// Render to raw bytes
const pdf = await client.render({
  markdown: "# Invoice\n\nDue: **$1,200**",
  template: "invoice",
  options: { format: "A4" },
});

// Render and get a signed download URL (expires in 15 min)
const { url, expires_at } = await client.renderSignedURL({
  markdown: "# Report",
});

// Upload a file (Node.js)
import { readFileSync } from "fs";
const content = readFileSync("invoice.md");
const pdf = await client.renderFile({
  file: new Blob([content]),
  filename: "invoice.md",
});

// Check usage
const usage = await client.usage();
console.log(`${usage.renders_used} / ${usage.renders_limit} renders used`);
```

## Error handling

```ts
import { DocRendersError } from "docrenders-sdk";

try {
  const pdf = await client.render({ markdown: "# Hello" });
} catch (err) {
  if (err instanceof DocRendersError) {
    console.error(err.code, err.message); // e.g. "quota_exceeded"
  }
}
```
