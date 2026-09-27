package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/charmbracelet/crush/internal/acp"
	"github.com/spf13/cobra"
)

var acpCmd = &cobra.Command{
	Use:   "acp",
	Short: "Start the ACP server on stdio",
	Long: `Start an Agent Client Protocol (ACP) v1 server speaking
newline-delimited JSON-RPC 2.0 over stdin/stdout.

Each session/prompt turn executes ` + "`crush run`" + ` non-interactively in the
session working directory and streams the assistant output to the client
as agent_message_chunk updates. Tool permissions are auto-approved by the
non-interactive runner, so only connect clients you trust with the
configured crush environment.

Session continuity across adapter restarts is provided through
session/load: ACP-to-crush session mappings persist in the state file
(override with CRUSH_ACP_STATE). The crush executable spawned for turns
defaults to this binary and can be overridden with CRUSH_ACP_BIN.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
		defer cancel()

		statePath, err := acp.DefaultStatePath()
		if err != nil {
			return fmt.Errorf("resolve ACP state path: %w", err)
		}
		store, err := acp.OpenStore(statePath)
		if err != nil {
			return fmt.Errorf("open ACP state store: %w", err)
		}

		server := acp.NewServer(os.Stdin, os.Stdout, acp.NewCrushRunner(), store, slog.Default())
		return server.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(acpCmd)
}
