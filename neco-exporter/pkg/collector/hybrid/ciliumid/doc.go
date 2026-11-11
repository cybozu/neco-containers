// Package ciliumid reports Cilium identities in both scopes.
//
//   - cluster: CiliumIdentity resources (identity_count, identity_info)
//   - node: identities of Pods on the node (identity_pods)
//
// Both scopes format security_labels with the same function, so that their metrics can be joined.
package ciliumid

const metricsPrefix = "ciliumid"
