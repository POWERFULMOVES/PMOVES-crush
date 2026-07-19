module.exports = {
  run: [
    {
      method: "shell.run",
      params: {
        message: [
          "echo '◇ crush-pmoves → bootstrapping PMOVES-aware Crush...'",
          "test -x \"{{args.cwd}}/../../../../../pmoves/scripts/crush-pmoves\" || { echo 'ERROR: PMOVES.AI repo root not found from Pinokio app directory' >&2; exit 1; }",
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
        message: "\"{{args.cwd}}/../../../../../pmoves/scripts/crush-pmoves\"",
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
