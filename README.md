# Stockyard Lasso

**Link shortener.** Shorten URLs, track clicks, custom slugs, vanity redirects. Your domain, your data. Single binary, no external dependencies.

Part of the [Stockyard](https://stockyard.dev) suite of self-hosted developer tools.

## Quick Start

```bash
curl -sfL https://stockyard.dev/install/lasso | sh
lasso
```

Dashboard at [http://localhost:8890/ui](http://localhost:8890/ui)

## Usage

```bash
# Shorten a URL
curl -X POST http://localhost:8890/api/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/very/long/path"}'
# → {"short_url":"http://localhost:8890/x4k9mn"}

# Custom slug
curl -X POST http://localhost:8890/api/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","slug":"my-brand"}'
# → http://localhost:8890/my-brand

# Click analytics
curl http://localhost:8890/api/links/{id}/stats
```

## Free vs Pro

| Feature | Free | Pro ($1.99/mo) |
|---------|------|----------------|
| Links | 25 | Unlimited |
| Click tracking | ✓ | ✓ |
| Custom slugs | ✓ | ✓ |
| Password protection | — | ✓ |
| Click retention | 7 days | 1 year |

## License

Apache 2.0 — see [LICENSE](LICENSE).
