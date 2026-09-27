# PMOVES.AI Integration Guide for PMOVES Danger Infra

## Integration Complete

The PMOVES.AI integration template has been applied to PMOVES Danger Infra.

## Next Steps

### 1. Customize Environment Variables

Edit the following files with your service-specific values:

- `env.shared` - Base environment configuration
- `env.tier-api` - API tier specific configuration
- `chit/secrets_manifest_v2.yaml` - Add your service's required secrets

### 2. Update Docker Compose

Add the PMOVES.AI environment anchor to your `docker-compose.yml`:

```yaml
services:
  danger-infra:
    <<: [*env-tier-api, *pmoves-healthcheck]
    # Your existing service configuration...
```

### 3. Integrate Health Check

Add the health check endpoint to your service:

```python
from pmoves_health import add_custom_check, get_health_status

@app.get("/healthz")
async def health_check():
    return await get_health_status()
```

### 4. Add Service Announcement

Add NATS service announcement to your startup:

```python
from pmoves_announcer import announce_service

@app.on_event("startup")
async def startup():
    await announce_service(
        slug="danger-infra",
        name="PMOVES Danger Infra",
        url=f"http://danger-infra:8065",
        port=8065,
        tier="api"
    )
```

### 5. Test Integration

```bash
# Test health check
curl http://localhost:8065/healthz

# Verify environment variables loaded
docker compose exec danger-infra env | grep PMOVES

# Verify NATS announcement
nats sub "services.announce.v1"
```

## Service Details

- **Name:** PMOVES Danger Infra
- **Slug:** danger-infra
- **Tier:** api
- **Port:** 8065
- **Health Check:** http://localhost:8065/healthz
- **NATS Enabled:** True
- **GPU Enabled:** False

## Files Created

- `env.shared` - Base PMOVES.AI environment
- `env.tier-api` - Tier-specific environment
- `chit/secrets_manifest_v2.yaml` - CHIT secrets configuration
- `pmoves_health/` - Health check module
- `pmoves_announcer/` - NATS service announcer
- `pmoves_registry/` - Service registry client
- `docker-compose.pmoves.yml` - PMOVES.AI YAML anchors

## Support

For questions or issues, see the PMOVES.AI documentation.

## ACP Server (crush acp) — added 2026-09-27

`crush acp` speaks Agent Client Protocol (ACP) v1 — newline-delimited
JSON-RPC 2.0 over stdio — so ACP clients (Zed, JetBrains, Spynel, the
agentclientprotocol registry ecosystem) can drive Crush headlessly.

- Methods: `initialize`, `session/new`, `session/load`, `session/resume`,
  `session/prompt` (async), `cancel`, `session/cancel` (notification);
  `permission/set_options` and `session/set_mode` are acknowledged no-ops
  because the underlying `crush run` turn auto-approves tool permissions.
- Each prompt turn execs `crush run --quiet [--session <uuid>] <prompt>`
  in the session cwd; stdout is streamed to the client as
  `session/update` / `agent_message_chunk` notifications (UTF-8 safe).
- Session continuity: ACP session id → crush session uuid mappings
  persist in the state file (`$XDG_CACHE_HOME/crush/acp-sessions.json`,
  override `CRUSH_ACP_STATE`); `session/load`/`session/resume` restore a
  conversation across adapter restarts. New crush sessions are discovered
  by diffing `crush session list --json` around the first turn of a
  session, which is why turns are serialized process-wide.
- Spawning: turns run the running binary by default (`os.Executable()`);
  override with `CRUSH_ACP_BIN`.
- Interop ground truth: PMOVES-spynel `internal/harness/acp.go` (first
  fleet consumer); registry entry `crush-acp` in POWERFULMOVES/PMOVES-registry.
- Files: `internal/acp/` (server, protocol, runner, state) +
  `internal/cmd/acp.go`; tests `internal/acp/server_test.go`.
