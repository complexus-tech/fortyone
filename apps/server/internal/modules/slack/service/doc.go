// Package slack owns the application behavior for the Slack integration.
// Provider transport details enter through the ingress capabilities and are
// translated into FortyOne service and repository contracts.
//
// The package is organized by reason to change:
//   - slack.go composes the service and serves integration read models.
//   - installation, OAuth, account linking, and member linking own lifecycle.
//   - request verification and the event, command, and interaction ingress
//     files own Slack's inbound protocols.
//   - event processor, recovery, assistant-delivery, and Work Object files own
//     durable event execution and provider-facing rich objects.
//   - actions, suggestions, submissions, and modal files own interactive work.
//   - receipts and outbound messages own provider responses and delivery.
//   - small payload, mapping, URL, selection, and source-context files contain
//     deterministic transformations shared by those capabilities.
//
// integration_contracts.go is the explicit compatibility seam for the
// package's pre-existing concrete cross-module types. New dependencies should
// be expressed as narrow caller-owned interfaces and wired by bootstrap.
package slack
