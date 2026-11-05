package httpx

import (
	"net/http"
	"time"
)

const DefaultTimeout = 15 * time.Second

func NoRedirects(client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: DefaultTimeout}
	}
	copied := *client
	copied.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &copied
}
