package gateway

import "sync"

type Resume struct {
	mu        sync.Mutex
	sessionID string
	sequence  *int64
}

func (r *Resume) Resumable() bool {
	sessionID, sequence := r.snapshot()
	return sessionID != "" && sequence != nil
}

func (r *Resume) Reset() {
	r.mu.Lock()
	r.sessionID = ""
	r.sequence = nil
	r.mu.Unlock()
}

func (r *Resume) observe(sequence *int64) {
	if sequence == nil {
		return
	}
	n := *sequence
	r.mu.Lock()
	r.sequence = &n
	r.mu.Unlock()
}

func (r *Resume) last() *int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sequence
}

func (r *Resume) snapshot() (string, *int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sessionID, r.sequence
}

func (r *Resume) start(sessionID string) {
	r.mu.Lock()
	r.sessionID = sessionID
	r.mu.Unlock()
}
