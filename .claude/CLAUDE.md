# gemini-cli-sdk-go

Go SDK for the Gemini CLI (`@google/gemini-cli`). Wraps the CLI as a subprocess and streams structured messages back over a channel.

## Build & Test

```bash
just build    # go build ./...
just test     # go test -race -count=1 ./...
just lint     # go vet ./...
just check    # fmt-check + lint + test + build
```

## Key Conventions

- **Module path**: `github.com/albertocavalcante/gemini-cli-sdk-go`
- **Public package**: `gemini/` — all exported types and functions
- **Internal transport**: `internal/transport/` — subprocess abstraction
- **Zero external deps**: Only Go stdlib
- **Tests**: Table-driven, MockTransport for all tests, no real CLI calls
- **Task runner**: `Justfile`
