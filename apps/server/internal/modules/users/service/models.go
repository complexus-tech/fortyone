package users

import usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"

const (
	MaximumOnboardingTourKeyRunes        = usersdomain.MaximumOnboardingTourKeyRunes
	MaximumOnboardingTourVersionRunes    = usersdomain.MaximumOnboardingTourVersionRunes
	MaximumOnboardingTourProgressIDs     = usersdomain.MaximumOnboardingTourProgressIDs
	MaximumOnboardingTourProgressIDRunes = usersdomain.MaximumOnboardingTourProgressIDRunes
	CoreOnboardingTourStatusActive       = usersdomain.OnboardingTourStatusActive
	CoreOnboardingTourStatusCompleted    = usersdomain.OnboardingTourStatusCompleted
	CoreOnboardingTourStatusSkipped      = usersdomain.OnboardingTourStatusSkipped
)

type CoreVerificationToken = usersdomain.VerificationToken
type NewVerificationToken = usersdomain.NewVerificationToken
type ConsumeVerificationTokenInput = usersdomain.ConsumeVerificationToken
type CoreUser = usersdomain.User
type CoreListUsersFilter = usersdomain.ListUsersFilter
type CoreUpdateUser = usersdomain.UpdateUser
type CoreWorkScheduleOverride = usersdomain.WorkScheduleOverride
type CoreNewUser = usersdomain.NewUser
type CoreExternalIdentityInput = usersdomain.ExternalIdentityInput
type CoreExternalIdentityResult = usersdomain.ExternalIdentityResult
type VerifiedSignInReactivation = usersdomain.VerifiedSignInReactivation
type CoreAutomationPreferences = usersdomain.AutomationPreferences
type CoreUpdateAutomationPreferences = usersdomain.UpdateAutomationPreferences
type CoreOnboardingTourStatus = usersdomain.OnboardingTourStatus
type CoreOnboardingTourScope = usersdomain.OnboardingTourScope
type CoreOnboardingTourProgress = usersdomain.OnboardingTourProgress
type CoreUpdateOnboardingTourProgress = usersdomain.UpdateOnboardingTourProgress
type CoreUserMemoryItem = usersdomain.UserMemoryItem
type NewUserMemoryItem = usersdomain.NewUserMemoryItem
type UserMemoryScope = usersdomain.UserMemoryScope
type UpdateUserMemoryItem = usersdomain.UpdateUserMemoryItem
