package acp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"unicode/utf8"
)

// Runner executes one non-interactive Crush turn. It is an interface so
// tests can drive the server without spawning real processes.
type Runner interface {
	// Run executes prompt against the Crush session named by sessionID
	// (empty creates a new session). Output chunks are streamed to
	// onChunk as they arrive. Run returns when the turn finishes.
	Run(ctx context.Context, cwd, sessionID, prompt string, onChunk func(string)) error
	// Sessions returns the Crush session UUIDs currently visible for cwd.
	Sessions(ctx context.Context, cwd string) ([]string, error)
}

type sessionListEntry struct {
	UUID string `json:"uuid"`
}

// crushRunner drives the crush CLI in non-interactive mode, one
// subprocess per prompt turn — the same integration class Spynel uses
// for Claude Code's print mode.
type crushRunner struct {
	bin string
}

// NewCrushRunner resolves the crush executable the adapter will spawn.
// CRUSH_ACP_BIN overrides; otherwise the running binary (version-matched)
// is preferred, falling back to "crush" on PATH.
func NewCrushRunner() Runner {
	bin := os.Getenv("CRUSH_ACP_BIN")
	if bin == "" {
		if exe, err := os.Executable(); err == nil {
			bin = exe
		} else {
			bin = "crush"
		}
	}
	return &crushRunner{bin: bin}
}

func (r *crushRunner) Sessions(ctx context.Context, cwd string) ([]string, error) {
	cmd := exec.CommandContext(ctx, r.bin, "session", "list", "--json")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list crush sessions: %w", err)
	}
	var entries []sessionListEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("decode crush session list: %w", err)
	}
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.UUID != "" {
			ids = append(ids, e.UUID)
		}
	}
	return ids, nil
}

func (r *crushRunner) Run(ctx context.Context, cwd, sessionID, prompt string, onChunk func(string)) error {
	args := []string{"run", "--quiet"}
	if sessionID != "" {
		args = append(args, "--session", sessionID)
	}
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, r.bin, args...)
	cmd.Dir = cwd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe crush stdout: %w", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start crush run: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		streamChunks(stdout, onChunk)
	}()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		tail := stderr.String()
		if len(tail) > 512 {
			tail = tail[len(tail)-512:]
		}
		if tail != "" {
			return fmt.Errorf("crush run failed: %w: %s", err, tail)
		}
		return fmt.Errorf("crush run failed: %w", err)
	}
	return nil
}

// streamChunks forwards reader output to onChunk without splitting a
// multi-byte UTF-8 sequence across chunks, so every chunk marshals into
// JSON without invalid-byte replacement.
func streamChunks(r io.Reader, onChunk func(string)) {
	if onChunk == nil {
		_, _ = io.Copy(io.Discard, r)
		return
	}
	buf := make([]byte, 8192)
	var carry []byte
	for {
		n, err := r.Read(buf)
		if n > 0 {
			data := append(carry, buf[:n]...)
			carry = nil
			cut := len(data)
			for cut > 0 && cut > len(data)-4 && !utf8.Valid(data[:cut]) {
				cut--
			}
			// Walk back at most one rune's continuation bytes: Valid
			// becomes true exactly at the boundary before the partial rune.
			if cut < len(data) {
				carry = append(carry, data[cut:]...)
			}
			if cut > 0 {
				onChunk(string(data[:cut]))
			}
		}
		if err != nil {
			if len(carry) > 0 {
				onChunk(string(carry))
			}
			return
		}
	}
}
