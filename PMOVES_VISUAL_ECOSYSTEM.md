# PMOVES-Crush Visual Ecosystem

> Submodule architecture for model registry, TUI tooling, terminal recording, and visual theming.

## Submodules

| Submodule | Upstream | Purpose |
|-----------|----------|---------|
| `pmoves-catwalk/` | `charmbracelet/catwalk` | Model registry with PMOVES providers (TensorZero, Z.AI, MiniMax) |
| `pmoves-gum/` | `charmbracelet/gum` (via `POWERFULMOVES/pmoves-gum`) | TUI scripting for fleet launchers, interactive prompts |
| `pmoves-vhs/` | `charmbracelet/vhs` (via `POWERFULMOVES/pmoves-vhs`) | Terminal recording for demos, documentation, Showtime content |

## pmoves-catwalk — Model Registry

Forked from `charmbracelet/catwalk` to add PMOVES-specific providers:

- **TensorZero** (`tensorzero.json`) — local LLM gateway routing (agent_zero function + embedding models)
- **Z.AI** (upstream) — GLM-5.2/5-Turbo coding plan at `api.z.ai/api/coding/paas/v4`
- **MiniMax** (upstream) — M2.7/M2.1 token plan models

### Usage

Build and run the catwalk server locally:
```bash
cd pmoves-catwalk
go run .  # serves on :8080
```

Crush's `go.mod` uses `charm.land/catwalk` — once `POWERFULMOVES/pmoves-catwalk` is published,
add a `replace` directive to route the import through the fork.

## pmoves-gum — TUI Scripting

Forked from `charmbracelet/gum` via `POWERFULMOVES/pmoves-gum`.

### Planned PMOVES Customizations

- **Agent theme integration** — gum styles driven by `agent_signatures.yaml` color palettes
- **Crush model picker** — `gum choose` populated from pmoves-catwalk provider list
- **NATS-aware commands** — `gum choose` options from NATS service discovery
- **PBnJ launcher menus** — replace ad-hoc pinokio.js UI logic with gum-powered interactive prompts

## pmoves-vhs — Terminal Recording

Forked from `charmbracelet/vhs` via `POWERFULMOVES/pmoves-vhs`.

### Planned PMOVES Customizations

- **PMOVES-themed tapes** — agent signature color palettes as VHS themes
- **Crush demo tapes** — `bringup.tape`, `cipher-smoke.tape`, `crush-pmoves.tape`
- **Showtime integration** — VHS output embedded in Showtime SSE stream
- **CI integration** — GitHub Action regenerates demo GIFs on release

## Theme Pipeline

```
agent_signatures.yaml
    ↓ generate.py
pmoves/design/themes/*.json
    ↓
pmoves-gum styles (terminal TUI)
pmoves-vhs themes (recording palettes)
PMOVES-crush themes.go (Go theme structs)
Showtime persona-theme.js (browser overlay)
```

## Integration with PMOVES.AI

These submodules live inside PMOVES-crush and are consumed by:
- **PMOVES-crush TUI** — catwalk types for model picker, gum for launcher UX
- **PBnJ launchers** — gum-powered menus in `pbnj/pinokio/api/pmoves-*`
- **Showtime API** — VHS-generated content in SSE events
- **Fleet scripts** — gum in `pmoves/scripts/` for interactive prompts
