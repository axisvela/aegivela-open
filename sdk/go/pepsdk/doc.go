// Package pepsdk is the Apache-2.0 AEGIVELA runtime-enforcement client:
//
//   - policy decisions: Authorize, SimulateDecision, VerifyDecision, DecisionJWKS;
//   - gateway/tool enforcement: GatewayConnect, ModuregisAuthorize, ToolInvoke;
//   - revocation: CheckRevocation, RevocationList;
//   - HTTP authorization middleware.
//
// Domain clients live in sibling packages: identitysdk (workload bindings +
// assertions + JWKS), grantsdk (execution grants + approvals + pre-auth +
// token-exchange), attestationsdk (attestations), governancesdk (policies, web
// resources, risk signals). The product-named root package
// github.com/axisrobo/aegivela-open/sdk/go/aegivela re-exports them all.
//
// Stability: the exported clients and their request/decision types are stable
// within a contract version. The embedded contract data lives in
// pepsdk/moduregiscontract and follows the contracts compatibility policy.
package pepsdk
