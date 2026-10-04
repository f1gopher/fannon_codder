# Fannon Codder

Cannon Fodder (Amiga, 1993) gameplay clone in Go and Ebitengine. Maps, art, and audio are original. The product name stays Fannon Codder.

Run from the repo root:

```
go run ./cmd/fannon
```

Flags: `-skip-title`, `-cover`, `-river`, `-hut`, `-hazards`, `-skidoo`.

| File | What it is |
| --- | --- |
| `docs/ARCHITECTURE.md` | How the game works now. Read it before changing code. |
| `docs/CHUNKS.md` | Work already done. |
| `docs/BACKLOG.md` | Work not done. |
| `controls.md` | Bindings. Generated from `internal/app/controls.go`. Do not edit by hand. |

`internal/sim` does not import Ebitengine or `internal/audio`. Check with `go test ./...` and `go build ./cmd/fannon`.
