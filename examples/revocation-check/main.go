// Example: use the AEGIVELA revocation recheck SDK client.
//
// Demonstrates three SLO classes (pre_dispatch, continuation, connection)
// and how each handles the negative cache differently.
//
// Build: cd examples/revocation-check && go build -o revocation-check .
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/axisrobo/aegivela-open/sdk/go/pepsdk"
)

func main() {
	client := pepsdk.NewClient(os.Getenv("AEGIVELA_URL"), nil)
	internalToken := os.Getenv("INTERNAL_TOKEN")
	if internalToken == "" {
		log.Fatal("INTERNAL_TOKEN is required (X-AEGIVELA-PEP value)")
	}

	selectors := []pepsdk.RevocationSelector{
		{Kind: "grant_jti", Value: "123e4567-e89b-42d3-a456-426614174000"},
		{Kind: "subject_ref", Value: "user-example"},
	}

	// Pre-dispatch: cache-free authoritative check. Use before dispatching work.
	err := client.CheckRevocation(context.Background(), internalToken,
		pepsdk.ClassPreDispatch, selectors, nil)
	fmt.Printf("Pre-dispatch:  %v\n", result(err))

	// Continuation: cache-free recheck at a declared boundary.
	err = client.CheckRevocation(context.Background(), internalToken,
		pepsdk.ClassContinuation, selectors, nil)
	fmt.Printf("Continuation:  %v\n", result(err))

	// Connection: bounded-TTL negative cache hit possible.
	err = client.CheckRevocation(context.Background(), internalToken,
		pepsdk.ClassConnection, selectors, nil)
	fmt.Printf("Connection:    %v\n", result(err))
}

func result(err error) string {
	switch err {
	case nil:
		return "clear"
	case pepsdk.ErrDenied:
		return "revoked"
	case pepsdk.ErrUnavailable:
		return "unavailable (fail closed)"
	case pepsdk.ErrUnauthenticated:
		return "unauthenticated"
	default:
		return fmt.Sprintf("error: %v", err)
	}
}
