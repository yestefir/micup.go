package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yestefir/micup.go/src/internal/httpx"
)

const (
	DefaultOrigin = "https://micup.gg"
	wellKnownPath = "/.well-known/fluxer"
	responseLimit = 1 << 20
)

type Endpoints struct {
	API     string
	Gateway string
}

func ParseOrigin(raw string) (*url.URL, error) {
	if raw == "" {
		raw = DefaultOrigin
	}
	origin, err := url.Parse(raw)
	if err != nil || origin.Host == "" || origin.User != nil || (origin.Scheme != "http" && origin.Scheme != "https") {
		return nil, errors.New("invalid Micup origin")
	}
	return origin, nil
}

func Discover(ctx context.Context, client *http.Client, origin *url.URL) (Endpoints, error) {
	base := strings.TrimRight(origin.String(), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+wellKnownPath, nil)
	if err != nil {
		return Endpoints{}, err
	}
	resp, err := httpx.NoRedirects(client).Do(req)
	if err != nil {
		return Endpoints{}, errors.New("Micup discovery is unavailable")
	}
	defer resp.Body.Close()
	var document struct {
		Endpoints struct {
			API     string `json:"api_public"`
			Gateway string `json:"gateway"`
		} `json:"endpoints"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(resp.Body, responseLimit)).Decode(&document) != nil {
		return Endpoints{}, errors.New("invalid Micup discovery response")
	}
	api, err := url.Parse(document.Endpoints.API)
	if err != nil || api.Host != origin.Host || api.Scheme != origin.Scheme {
		return Endpoints{}, errors.New("Micup API origin mismatch")
	}
	gateway, err := url.Parse(document.Endpoints.Gateway)
	if err != nil || gateway.Host != origin.Host || (gateway.Scheme != "ws" && gateway.Scheme != "wss") || (origin.Scheme == "https" && gateway.Scheme != "wss") {
		return Endpoints{}, errors.New("invalid Micup Gateway endpoint")
	}
	return Endpoints{API: strings.TrimRight(api.String(), "/"), Gateway: gateway.String()}, nil
}
