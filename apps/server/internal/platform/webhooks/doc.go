// Package webhooks owns the provider-neutral inbound webhook control plane.
//
// Provider adapters still own signature algorithms, payload schemas, event
// normalization, and installation lookups. This package owns the invariants
// every adapter shares: bounded signed requests, immutable delivery identity,
// encrypted durable receipts, deduplication, queue-by-ID, leases, recovery,
// and payload retention.
package webhooks
