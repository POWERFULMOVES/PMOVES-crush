module.exports = {
  run: [
    {
      method: "shell.run",
      params: {
        message: [
          "echo '◇ crush-pmoves → bootstrapping PMOVES-aware Crush...'",
          "make -C \"{{args.cwd}}/../../../pmoves\" crush-bootstrap 2>/dev/null || echo 'Bootstrap skipped (may need secrets-funnel first)'",
          "echo ''",
          "echo '◇ Launching Crush...'",
          "echo '  Model: GLM-5.2 (Z.AI Coding Plan)'",
          "echo '  MCP: pmoves-mini, pmoves-nats-fleet, pmoves-supabase, pmoves-cipher'",
          "echo '  LSP: gopls, pyright, typescript-language-server'",
          "echo ''"
        ]
      }
    },
    {
      method: "shell.run",
      params: {
        message: "crush",
        on: [{
          event: "/(http:\\/\\/[0-9.:]+)/",
          done: true
        }]
      }
    },
    {
      method: "local.set",
      params: {
        url: "{{input.event[1]}}"
      }
    }
  ]
}
