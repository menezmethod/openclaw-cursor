package streaming

import (
	"bufio"
	"encoding/json"
	"io"
)

// Scanner reads NDJSON lines from cursor-agent stdout.
type Scanner struct {
	scanner *bufio.Scanner
}

// NewScanner creates a scanner for the given reader.
func NewScanner(r io.Reader) *Scanner {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64*1024), 1024*1024) // 64KB initial, 1MB max line
	return &Scanner{scanner: s}
}

// Scan advances to the next line. Returns false when no more input.
func (s *Scanner) Scan() bool {
	return s.scanner.Scan()
}

// Event parses the current line as a StreamEvent.
// Returns nil for empty lines or malformed JSON (caller should skip).
func (s *Scanner) Event() (*StreamEvent, error) {
	line := s.scanner.Bytes()
	if len(line) == 0 {
		return nil, nil
	}
	var e StreamEvent
	if err := json.Unmarshal(line, &e); err != nil {
		return nil, err
	}
	if e.Type == "" {
		return nil, nil
	}
	return &e, nil
}

// Err returns any scan error.
func (s *Scanner) Err() error {
	return s.scanner.Err()
}
