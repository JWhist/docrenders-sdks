# pdfgen-go

Official Go SDK for the [DocRenders API](https://docrenders.com).

## Installation

```bash
go get github.com/JWhist/docrenders-sdks/go
```

## Usage

```go
import (
    "context"
    "os"

    pdfgen "github.com/JWhist/docrenders-sdks/go"
)

client := docrenders.NewClient(os.Getenv("DOCRENDERS_API_KEY"))
ctx := context.Background()

// Render to raw bytes
pdf, err := client.Render(ctx, docrenders.RenderRequest{
    Markdown: "# Invoice\n\nDue: **$1,200**",
    Template: "invoice",
    Options:  docrenders.RenderOptions{Format: "A4"},
})

// Render and get a signed download URL (expires in 15 min)
result, err := client.RenderSignedURL(ctx, docrenders.RenderRequest{
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
