# PDFGen SDKs

Official client libraries for the PDFGen API — convert Markdown or HTML to production-ready PDFs with a single API call.

| SDK | Directory | Package |
|-----|-----------|---------|
| Go | [`/go`](./go) | `github.com/JWhist/pdfgen-sdks/go` |
| JavaScript / TypeScript | [`/js`](./js) | `pdfgen-sdk` (npm) |
| Python | [`/python`](./python) | `pdfgen-sdk` (PyPI) |

## Quick start

### Go

```go
import pdfgen "github.com/JWhist/pdfgen-sdks/go"

client := pdfgen.NewClient("pdg_live_YOUR_API_KEY")
pdf, err := client.Render(ctx, pdfgen.RenderRequest{
    Markdown: "# Invoice\n\nDue: **$1,200**",
})
```

### JavaScript / TypeScript

```ts
import PDFGenClient from "pdfgen-sdk";

const client = new PDFGenClient("pdg_live_YOUR_API_KEY");
const pdf = await client.render({ markdown: "# Invoice\n\nDue: **$1,200**" });
```

### Python

```python
from pdfgen import PDFGenClient, RenderRequest

client = PDFGenClient("pdg_live_YOUR_API_KEY")
pdf = client.render(RenderRequest(markdown="# Invoice\n\nDue: **$1,200**"))
```

## API reference

See the [PDFGen API docs](https://pdfgen-api.netlify.app/docs.html) for the full reference.
