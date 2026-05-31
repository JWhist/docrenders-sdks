# pdfgen-go

Official Go SDK for the [PDFGen API](https://pdfgen-api.netlify.app).

## Installation

```bash
go get github.com/JWhist/pdfgen-sdks/go
```

## Usage

```go
import (
    "context"
    "os"

    pdfgen "github.com/JWhist/pdfgen-sdks/go"
)

client := pdfgen.NewClient(os.Getenv("PDFGEN_API_KEY"))
ctx := context.Background()

// Render to raw bytes
pdf, err := client.Render(ctx, pdfgen.RenderRequest{
    Markdown: "# Invoice\n\nDue: **$1,200**",
    Template: "invoice",
    Options:  pdfgen.RenderOptions{Format: "A4"},
})

// Render and get a signed download URL (expires in 15 min)
result, err := client.RenderSignedURL(ctx, pdfgen.RenderRequest{
    Markdown: "# Report",
})
fmt.Println(result.URL)

// Upload a file
content, _ := os.ReadFile("invoice.md")
pdf, err = client.RenderFile(ctx, pdfgen.RenderFileRequest{
    Filename: "invoice.md",
    Content:  content,
})

// Check usage
usage, err := client.Usage(ctx)
fmt.Printf("%d / %d renders used\n", usage.RendersUsed, usage.RendersLimit)
```
