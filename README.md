# DocRenders SDKs

Official client libraries for the DocRenders API — convert Markdown or HTML to production-ready PDFs with a single API call.

| SDK | Directory | Package |
|-----|-----------|---------|
| Go | [`/go`](./go) | `github.com/JWhist/docrenders-sdks/go` |
| JavaScript / TypeScript | [`/js`](./js) | `docrenders-sdk` (npm) |
| Python | [`/python`](./python) | `docrenders-sdk` (PyPI) |

## Quick start

### Go

```go
import docrenders "github.com/JWhist/docrenders-sdks/go"

client := docrenders.NewClient("dcr_live_YOUR_API_KEY")
pdf, err := client.Render(ctx, docrenders.RenderRequest{
    Markdown: "# Invoice\n\nDue: **$1,200**",
})
```

### JavaScript / TypeScript

```ts
import DocRendersClient from "docrenders-sdk";

const client = new DocRendersClient("dcr_live_YOUR_API_KEY");
const pdf = await client.render({ markdown: "# Invoice\n\nDue: **$1,200**" });
```

### Python

```python
from docrenders import DocRendersClient, RenderRequest

client = DocRendersClient("dcr_live_YOUR_API_KEY")
pdf = client.render(RenderRequest(markdown="# Invoice\n\nDue: **$1,200**"))
```

## API reference

See the [DocRenders API docs](https://www.docrenders.com/docs.html) for the full reference.
