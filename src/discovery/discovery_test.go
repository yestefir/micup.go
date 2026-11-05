package discovery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serve(t *testing.T, api, gateway func(base string) string) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wellKnownPath {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": map[string]string{"api_public": api(server.URL), "gateway": gateway(server.URL)}})
	}))
	t.Cleanup(server.Close)
	return server
}

func wsURL(base string) string { return "ws" + base[4:] + "/gateway" }

func TestDiscoverReturnsEndpoints(t *testing.T) {
	server := serve(t, func(base string) string { return base + "/api/" }, wsURL)
	origin, err := ParseOrigin(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	endpoints, err := Discover(context.Background(), http.DefaultClient, origin)
	if err != nil {
		t.Fatal(err)
	}
	if endpoints.API != server.URL+"/api" || endpoints.Gateway != wsURL(server.URL) {
		t.Fatalf("unexpected endpoints: %+v", endpoints)
	}
}

func TestDiscoverRejectsForeignAPI(t *testing.T) {
	server := serve(t, func(string) string { return "http://attacker.test/api" }, wsURL)
	origin, _ := ParseOrigin(server.URL)
	if _, err := Discover(context.Background(), http.DefaultClient, origin); err == nil {
		t.Fatal("foreign API origin accepted")
	}
}

func TestDiscoverRejectsForeignGateway(t *testing.T) {
	server := serve(t, func(base string) string { return base + "/api" }, func(string) string { return "ws://attacker.test/gateway" })
	origin, _ := ParseOrigin(server.URL)
	if _, err := Discover(context.Background(), http.DefaultClient, origin); err == nil {
		t.Fatal("foreign Gateway origin accepted")
	}
}

func TestParseOriginRejectsCredentialsAndSchemes(t *testing.T) {
	for _, raw := range []string{"ftp://micup.gg", "https://user:pass@micup.gg", "micup.gg"} {
		if _, err := ParseOrigin(raw); err == nil {
			t.Fatalf("origin accepted: %s", raw)
		}
	}
	if origin, err := ParseOrigin(""); err != nil || origin.String() != DefaultOrigin {
		t.Fatalf("default origin not used: %v %v", origin, err)
	}
}
