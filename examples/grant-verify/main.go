// Example: verify an AEGIVELA execution grant.
//
// Demonstrates grant signature verification, claim inspection,
// revocation recheck, and scope/audience/expiry validation.
//
// Build: cd examples/grant-verify && go build -o grant-verify .
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/axisrobo/aegivela-open/sdk/go/pepsdk"
)

func main() {
	grantToken := os.Getenv("GRANT_TOKEN")
	if grantToken == "" {
		log.Fatal("GRANT_TOKEN is required")
	}

	// Verify the grant through the AEGIVELA server.
	// In production, the Gateway PEP verifies grants locally with Ed25519 keys;
	// this example shows the SDK-based verification via the server.
	client := pepsdk.NewClient(os.Getenv("AEGIVELA_URL"), nil)

	// Perform a policy evaluation with the grant as parent authority.
	// This proves the grant is valid, unrevoked, and within scope constraints.
	result, err := client.Authorize(context.Background(), pepsdk.Route{
		Action:    "read",
		Kind:      "api:example",
		Reference: "demo-resource",
		Scope:     []string{"data:read"},
		Audience:  "example-client",
		Modes:     []pepsdk.Mode{pepsdk.ModeDelegatedAPI},
	}, pepsdk.AuthorizationInput{
		Mode:                pepsdk.ModeDelegatedAPI,
		BearerToken:         os.Getenv("BEARER_TOKEN"),
		ParentExecutionGrant: grantToken,
	})
	if err != nil {
		log.Fatalf("Grant verification failed: %v", err)
	}

	fmt.Printf("Grant verified successfully:\n")
	fmt.Printf("  Decision ID:   %s\n", result.DecisionID)
	fmt.Printf("  Policy Version: %s\n", result.PolicyVersion)
	fmt.Printf("  Outcome:        %s\n", result.Outcome)
	fmt.Printf("  Evidence Refs:  %v\n", result.EvidenceRefs)

	// Perform a continuation recheck for a long-running operation.
	// pre_dispatch → cache-free authoritative check.
	err = client.CheckRevocation(context.Background(),
		os.Getenv("INTERNAL_TOKEN"),
		pepsdk.ClassContinuation,
		[]pepsdk.RevocationSelector{
			{Kind: "grant_jti", Value: "from-verified-grant"},
		}, nil)
	if err != nil {
		fmt.Printf("\nContinuation recheck: REVOKED or UNAVAILABLE → halt\n")
	} else {
		fmt.Printf("\nContinuation recheck: clear\n")
	}

	// Graceful expiration window
	_ = time.Now().Add(5 * time.Minute)
	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("\nFull result:\n%s\n", output)
}
