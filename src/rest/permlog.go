package rest

import (
	"strings"
	"sync"
	"time"
)

const (
	permissionLogWindow = time.Minute
	permissionLogLimit  = 512
)

type permissionLog struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

func (p *permissionLog) allow(key string, now time.Time) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if last, ok := p.seen[key]; ok && now.Sub(last) < permissionLogWindow {
		return false
	}
	if p.seen == nil {
		p.seen = map[string]time.Time{}
	}
	if len(p.seen) >= permissionLogLimit {
		for k, at := range p.seen {
			if now.Sub(at) >= permissionLogWindow {
				delete(p.seen, k)
			}
		}
		if len(p.seen) >= permissionLogLimit {
			clear(p.seen)
		}
	}
	p.seen[key] = now
	return true
}

func isPermissionError(err *APIError) bool {
	return err.StatusCode == 403 && (err.Code == CodeMissingPermissions || err.Code == CodeMissingAccess)
}

func (c *Client) notePermissionDenied(r *request, err *APIError) {
	if c.logger == nil || !isPermissionError(err) {
		return
	}
	if !c.permissions.allow(r.method+" "+r.path, time.Now()) {
		return
	}
	attrs := []any{"method", r.method, "path", r.path, "code", err.Code}
	if guildID := guildOf(r.path); guildID != "" {
		attrs = append(attrs, "guild", guildID)
	}
	if err.Message != "" {
		attrs = append(attrs, "message", err.Message)
	}
	c.logger.Warn("bot lacks permission for this request", attrs...)
}

func guildOf(requestPath string) string {
	parts := strings.Split(strings.Trim(requestPath, "/"), "/")
	if len(parts) >= 2 && parts[0] == "guilds" {
		return parts[1]
	}
	return ""
}
