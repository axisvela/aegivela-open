// Example: protect an HTTP API with the AEGIVELA PEP SDK.
//
// This demo shows three authorization modes on a single server:
//   GET  /api/public   -- human_web (browser user with enterprise IdP access token)
//   POST /api/system   -- system_api (service-to-service with workload assertion)
//   GET  /api/delegated -- delegated_api (system acting on behalf of user)
//
// Build: cd examples/web-api-pep && go build -o pep-demo .
// Run:   AEGIVELA_URL=http://localhost:8080 ./pep-demo
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/axisrobo/aegivela-open/sdk/go/pepsdk"
)

func main() {
	aegivelaURL := os.Getenv("AEGIVELA_URL")
	if aegivelaURL == "" {
		aegivelaURL = "http://localhost:8080"
	}

	client := pepsdk.NewClient(aegivelaURL, http.DefaultClient)

	routes := []struct {
		path  string
		route pepsdk.Route
		mode  pepsdk.Mode
	}{
		{
			"/api/public",
			pepsdk.Route{Action: "read", Kind: "api:public", Reference: "dashboard", Scope: []string{"public:read"}, Audience: "my-product-api"},
			pepsdk.ModeHumanWeb,
		},
		{
			"/api/system",
			pepsdk.Route{Action: "write", Kind: "api:system", Reference: "ingestion", Scope: []string{"system:write"}, Audience: "my-product-api"},
			pepsdk.ModeSystemAPI,
		},
		{
			"/api/delegated",
			pepsdk.Route{Action: "read", Kind: "api:delegated", Reference: "reports", Scope: []string{"reports:read"}, Audience: "my-product-api"},
			pepsdk.ModeDelegatedAPI,
		},
	}

	mux := http.NewServeMux()
	for _, rt := range routes {
		path := rt.path
		auth := pepsdk.NewAuthorizeMiddleware(client, pepsdk.WithRoute(rt.route), pepsdk.WithMode(rt.mode))
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"path": path, "status": "authorized"})
		})
		mux.Handle(path, auth.Middleware(handler))
	}

	addr := ":3000"
	log.Printf("PEP demo listening on %s, AEGIVELA at %s", addr, aegivelaURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
