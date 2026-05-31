# pdfgen-sdk (Python)

Official Python SDK for the [PDFGen API](https://pdfgen-api.netlify.app). No dependencies — uses the standard library only.

## Installation

```bash
pip install pdfgen-sdk
```

## Usage

```python
import os
from pdfgen import PDFGenClient, RenderRequest, RenderFileRequest, RenderOptions

client = PDFGenClient(os.environ["PDFGEN_API_KEY"])

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
from pdfgen import PDFGenClient, PDFGenError, RenderRequest

try:
    pdf = client.render(RenderRequest(markdown="# Hello"))
except PDFGenError as e:
    print(e.code, str(e))  # e.g. "quota_exceeded"
```
