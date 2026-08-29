package codehost

import "testing"

func TestWorkItemReconciliationSuppressesOnlyMatchingFortyOneEcho(t *testing.T) {
	t.Parallel()
	windowsRevision := WorkItemRevision(WorkItemSnapshot{
		Title: "Ship integration",
		Body:  "Line one\r\nLine two",
		State: WorkItemStateClosed,
	})
	unixRevision := WorkItemRevision(WorkItemSnapshot{
		Title: "Ship integration",
		Body:  "Line one\nLine two",
		State: WorkItemStateClosed,
	})
	if windowsRevision != unixRevision {
		t.Fatal("WorkItemRevision() changed across newline conventions")
	}
	if !ShouldSuppressEcho(ReconciliationState{
		LastOrigin: SyncOriginFortyOne, LastRevision: windowsRevision,
	}, unixRevision) {
		t.Fatal("ShouldSuppressEcho() = false for matching FortyOne revision")
	}
	if ShouldSuppressEcho(ReconciliationState{
		LastOrigin: SyncOrigin("github"), LastRevision: windowsRevision,
	}, unixRevision) {
		t.Fatal("ShouldSuppressEcho() = true for provider-originated revision")
	}
}
