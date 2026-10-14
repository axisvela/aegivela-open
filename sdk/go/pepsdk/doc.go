// Package pepsdk is the Apache-2.0 Go client SDK for AEGIVELA. It covers:
//
//   - policy enforcement: Authorize (policy decision), GatewayConnect,
//     ModuregisAuthorize, and CheckRevocation, plus the HTTP middleware;
//   - workload identity: IssueWorkloadAssertion, CreateWorkloadBinding,
//     WorkloadAssertionJWKS;
//   - execution grants: IssueGrant, VerifyGrant, GrantJWKS;
//   - attestations: IssueAttestation, VerifyAttestation, AttestationJWKS.
//
// The product-named root package github.com/axisrobo/aegivela-open/sdk/go/aegivela
// re-exports this surface for convenience.
//
// Stability: the exported clients and their request/decision types are stable
// within a contract version. The embedded contract data lives in
// pepsdk/moduregiscontract and follows the contracts compatibility policy.
package pepsdk
