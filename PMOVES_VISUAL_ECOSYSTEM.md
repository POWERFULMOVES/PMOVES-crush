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

The `pmoves-catwalk` gitlink records the PMOVES fork alongside Crush. The
optional local `replace charm.land/catwalk => ./pmoves-catwalk` directive is
documented in `go.mod`, but remains disabled because normal CI does not
initialize submodules. The built binary therefore uses the pinned upstream
Catwalk module until that replace is explicitly enabled.

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

```text
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

### Current vs Planned

| Component | Current Status | Planned Integration |
|-----------|---------------|---------------------|
| pmoves-catwalk | Recorded submodule; local replace documented but disabled | Validate TensorZero fork and enable it in submodule-aware builds |
| pmoves-gum | Submodule (clean mirror) + `gum-pmoves` wrapper script | Go-native theme injection |
| pmoves-vhs | Submodule (clean mirror) + 3 demo tapes | CI-integrated demo regeneration |
| Showtime | `showtime-status` CLI script | Go-native TUI status bar |
| PBnJ | `pmoves-crush` Pinokio launcher | Full PBnJ app migration |
