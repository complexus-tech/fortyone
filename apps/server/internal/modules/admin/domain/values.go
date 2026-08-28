package admindomain

import (
	"fmt"
	"strings"
)

type WorkspaceStatus string

const (
	WorkspaceStatusAll      WorkspaceStatus = ""
	WorkspaceStatusActive   WorkspaceStatus = "active"
	WorkspaceStatusTrialing WorkspaceStatus = "trialing"
	WorkspaceStatusExpired  WorkspaceStatus = "expired"
	WorkspaceStatusExpiring WorkspaceStatus = "expiring"
	WorkspaceStatusPaid     WorkspaceStatus = "paid"
	WorkspaceStatusPastDue  WorkspaceStatus = "past_due"
	WorkspaceStatusDeleted  WorkspaceStatus = "deleted"
)

func ParseWorkspaceStatus(value string) (WorkspaceStatus, error) {
	status := WorkspaceStatus(strings.ToLower(strings.TrimSpace(value)))
	switch status {
	case WorkspaceStatusAll, WorkspaceStatusActive, WorkspaceStatusTrialing,
		WorkspaceStatusExpired, WorkspaceStatusExpiring, WorkspaceStatusPaid,
		WorkspaceStatusPastDue, WorkspaceStatusDeleted:
		return status, nil
	default:
		return WorkspaceStatusAll, fmt.Errorf("%w: workspace status %q", ErrInvalidFilter, value)
	}
}

type TargetType string

const (
	TargetAny          TargetType = ""
	TargetWorkspace    TargetType = "workspace"
	TargetUser         TargetType = "user"
	TargetSubscription TargetType = "subscription"
	TargetSystem       TargetType = "system"
)

func ParseTargetType(value string) (TargetType, error) {
	target := TargetType(strings.ToLower(strings.TrimSpace(value)))
	switch target {
	case TargetAny, TargetWorkspace, TargetUser, TargetSubscription, TargetSystem:
		return target, nil
	default:
		return TargetAny, fmt.Errorf("%w: target type %q", ErrInvalidFilter, value)
	}
}

func ParseNoteTargetType(value string) (TargetType, error) {
	target, err := ParseTargetType(value)
	if err != nil {
		return TargetAny, err
	}
	if target != TargetWorkspace && target != TargetUser {
		return TargetAny, fmt.Errorf("%w: note target type %q", ErrInvalidAction, value)
	}
	return target, nil
}

type AuditAction string

type LoginReactivationPolicy string

const (
	LoginReactivationVerifiedSignIn LoginReactivationPolicy = "verified_sign_in"
	LoginReactivationAdminOnly      LoginReactivationPolicy = "admin_only"
	LoginReactivationLegacyReview   LoginReactivationPolicy = "legacy_admin_review"
)

const (
	AuditActionAny                      AuditAction = ""
	AuditWorkspaceTrialUpdated          AuditAction = "workspace.trial_updated"
	AuditWorkspaceDeleted               AuditAction = "workspace.deleted"
	AuditWorkspaceRestored              AuditAction = "workspace.restored"
	AuditUserActivated                  AuditAction = "user.activated"
	AuditUserDeactivated                AuditAction = "user.deactivated"
	AuditUserInternalGranted            AuditAction = "user.internal_granted"
	AuditUserInternalRevoked            AuditAction = "user.internal_revoked"
	AuditUserStateReviewed              AuditAction = "user.state_reviewed"
	AuditUserReactivationPolicyChanged  AuditAction = "user.reactivation_policy_changed"
	AuditUserSessionRevocationRequested AuditAction = "user.session_revocation_requested"
	AuditAdminNoteCreated               AuditAction = "admin_note.created"
	AuditSubscriptionSyncRequested      AuditAction = "subscription.sync_requested"
	AuditSubscriptionSynced             AuditAction = "subscription.synced"
	AuditSubscriptionSyncFailed         AuditAction = "subscription.sync_failed"
)

func ParseAuditAction(value string) (AuditAction, error) {
	action := AuditAction(strings.ToLower(strings.TrimSpace(value)))
	switch action {
	case AuditActionAny, AuditWorkspaceTrialUpdated, AuditWorkspaceDeleted,
		AuditWorkspaceRestored, AuditUserActivated, AuditUserDeactivated,
		AuditUserInternalGranted, AuditUserInternalRevoked, AuditUserStateReviewed,
		AuditUserReactivationPolicyChanged,
		AuditUserSessionRevocationRequested, AuditAdminNoteCreated,
		AuditSubscriptionSyncRequested, AuditSubscriptionSynced,
		AuditSubscriptionSyncFailed:
		return action, nil
	default:
		return AuditActionAny, fmt.Errorf("%w: audit action %q", ErrInvalidFilter, value)
	}
}

type SubscriptionSyncOutcome string

const (
	SubscriptionSyncSucceeded SubscriptionSyncOutcome = "succeeded"
	SubscriptionSyncFailed    SubscriptionSyncOutcome = "failed"
)
