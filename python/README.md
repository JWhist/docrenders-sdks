# docrenders-sdk (Python)

Official Python SDK for the [DocRenders API](https://www.docrenders.com). No dependencies — uses the standard library only.

## Installation

```bash
pip install docrenders-sdk
```

## Usage

```python
import os
from docrenders import DocRendersClient, RenderRequest, RenderFileRequest, RenderOptions

client = DocRendersClient(os.environ["PDFGEN_API_KEY"])

# Render to raw bytes
pdf = client.render(RenderRequest(
    markdown="# Invoice\n\nDue: **$1,200**",
    template="invoice",
    options=RenderOptions(format="A4"),
))

# Render and get a signed download URL (expires in 15 min)
result = client.render_signed_url(RenderRequest(markdown="# Report"))
print(result.url)

# Upload a file
with open("invoice.md", "rb") as f:
    pdf = client.render_file(RenderFileRequest(
        filename="invoice.md",
        content=f.read(),
    ))

# Check usage
usage = client.usage()
print(f"{usage.renders_used} / {usage.renders_limit} renders used")
```

## Error handling

```python
from docrenders import DocRendersClient, DocRendersError, RenderRequest

try:
    pdf = client.render(RenderRequest(markdown="# Hello"))
except DocRendersError as e:
    print(e.code, str(e))  # e.g. "quota_exceeded"
```
