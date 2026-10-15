//go:build conformance

package pepsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestEvaluateV2RequestsConformToPublishedSchema(t *testing.T) {
	path := filepath.Join("..", "..", "..", "contracts", "policy", "v2.0", "policy-decision.schema.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	const id = "https://aegivela.io/contracts/policy/v2.0/policy-decision.schema.json"
	if document["$id"] != id {
		t.Fatalf("schema ID = %v", document["$id"])
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(id, document); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(id + "#/$defs/policyDecisionRequest")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []Mode{ModeHumanWeb, ModeSystemAPI, ModeDelegatedAPI, ModeServiceAgentAPI, ModeTwinAgentAPI} {
		t.Run(string(mode), func(t *testing.T) {
			input := v2AgentInput(mode)
			if mode == ModeHumanWeb || mode == ModeSystemAPI || mode == ModeDelegatedAPI {
				input.ParentAuthority = nil
				input.Principal = &PolicyPrincipalV2{Namespace: "ns", ActorRef: "actor"}
				if mode == ModeSystemAPI || mode == ModeDelegatedAPI {
					input.Principal.ClientID, input.Principal.WorkloadRef = "client", "workload"
				}
				if mode == ModeHumanWeb || mode == ModeDelegatedAPI {
					input.Principal.SubjectRef, input.BearerToken = "subject", "bearer"
				}
				if mode == ModeHumanWeb {
					input.WorkloadAssertion = ""
				}
				if mode == ModeSystemAPI {
					input.Principal.OrganizationAuthorityRoot = "org"
				}
				if mode == ModeDelegatedAPI {
					input.ParentExecutionGrant = "signed.parent.grant"
				}
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var wire any
				if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
					t.Error(err)
				}
				if err := schema.Validate(wire); err != nil {
					t.Errorf("SDK violates public v2 schema: %v", err)
				}
				_, _ = w.Write([]byte(`{"apiVersion":"aegivela.io/v2.0","decisionId":"d1","outcome":"allow","policyVersion":"v1","evidenceRefs":["e"]}`))
			}))
			defer srv.Close()
			if _, err := NewClient(srv.URL, srv.Client()).Evaluate(context.Background(), v2AgentRoute(), input); err != nil {
				t.Fatal(err)
			}
		})
	}
}
