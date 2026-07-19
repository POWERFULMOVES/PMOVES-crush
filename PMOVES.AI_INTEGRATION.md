# PMOVES.AI Integration Contract

PMOVES-Crush tracks the current Charmbracelet Crush runtime and layers optional
PMOVES.AI integration around it. The Go application does not require PMOVES
services to start; integrations must verify configured endpoints before use.

## Runtime integration

- The packaged command and release artifacts are named `pmoves-crush`.
- The live coder prompt recognizes TensorZero, Hi-RAG, Agent Zero, and NATS when
  project configuration confirms they are available.
- `pmoves_announcer`, `pmoves_health`, and `pmoves_registry` are reusable Python
  helpers for services launched alongside PMOVES-Crush. They are not imported by
  the Go binary.
- `docker-compose.pmoves.yml` supplies API and data-tier anchors. It is a
  composition fragment, not a standalone PMOVES stack.

## Environment

Source the shared defaults before the appropriate tier file:

```bash
source env.shared
source env.tier-api.sh  # or env.tier-data.sh
```

Secrets default to empty values. Populate them through the PMOVES secrets funnel
or the deployment environment; do not commit credentials. The CHIT manifest at
`chit/secrets_manifest_v2.yaml` declares the expected secret names.

## Optional service helpers

Health endpoint:

```python
from pmoves_health import get_health_status

@app.get("/healthz")
async def health_check():
    return await get_health_status()
```

NATS announcement:

```python
from pmoves_announcer import announce_service

@app.on_event("startup")
async def startup():
    await announce_service(
        slug="pmoves-crush",
        name="PMOVES-Crush",
        url="http://pmoves-crush:8080",
        port=8080,
        tier="agent",
    )
```

## Validation

```bash
go build ./...
go test ./internal/agent ./internal/cmd ./internal/config
docker compose -f docker-compose.pmoves.yml config --quiet
```

The optional visual submodules and launcher scripts are documented in
`PMOVES_VISUAL_ECOSYSTEM.md`.
