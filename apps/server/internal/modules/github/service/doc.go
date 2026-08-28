// Package github owns GitHub integration use cases and provider translation.
//
// The package is organized by reason to change: installation and settings;
// story, issue, pull-request, review, check, and comment synchronization;
// OAuth/user linking; provider API translation; and durable webhook ingress,
// processing, and recovery. codehost_adapter.go is the provider-neutral
// capability facade. github.go is the explicit legacy compatibility seam; new
// dependencies belong behind narrow caller-owned interfaces.
package github
