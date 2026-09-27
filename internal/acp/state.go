package acp

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// storedSession persists the mapping between the ACP session ID handed to
// the client and the underlying Crush session ID, so conversations survive
// adapter restarts via session/load.
type storedSession struct {
	ACPSessionID   string `json:"acpSessionId"`
	CrushSessionID string `json:"crushSessionId,omitempty"`
	Cwd            string `json:"cwd"`
}

// Store keeps ACP-to-Crush session mappings in a JSON file.
type Store struct {
	mu       sync.Mutex
	path     string
	sessions map[string]*storedSession
}

// DefaultStatePath returns the default session-map location, overridable
// with CRUSH_ACP_STATE for tests and sandboxed deployments.
func DefaultStatePath() (string, error) {
	if p := os.Getenv("CRUSH_ACP_STATE"); p != "" {
		return p, nil
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "crush", "acp-sessions.json"), nil
}

// OpenStore loads (or creates) the session store at path.
func OpenStore(path string) (*Store, error) {
	s := &Store{path: path, sessions: map[string]*storedSession{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &s.sessions); err != nil {
		return nil, err
	}
	if s.sessions == nil {
		s.sessions = map[string]*storedSession{}
	}
	return s, nil
}

func (s *Store) get(acpID string) (*storedSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[acpID]
	return sess, ok
}

func (s *Store) put(sess *storedSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.ACPSessionID] = sess
	data, err := json.MarshalIndent(s.sessions, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
