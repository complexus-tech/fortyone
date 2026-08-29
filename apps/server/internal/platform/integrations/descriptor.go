package integrations

import "time"

// ProviderKey is a stable machine identifier persisted in installation,
// credential, inbox, and audit records. Display names are never identifiers.
type ProviderKey string

// Family groups providers that share control-plane behavior without claiming
// that their native resource models are identical.
type Family string

const (
	FamilyCodeHost        Family = "code_host"
	FamilyMessaging       Family = "messaging"
	FamilyCalendar        Family = "calendar"
	FamilySupportFeedback Family = "support_feedback"
	FamilyDesignContext   Family = "design_context"
)

// CapabilityKey identifies one narrow provider behavior contract.
type CapabilityKey string

// Capability declares the contract version implemented by a provider. Major
// versions are intentionally explicit so incompatible adapter contracts cannot
// be selected silently.
type Capability struct {
	Key          CapabilityKey
	MajorVersion uint16
}

// AuthStrategy describes a provider authentication/install flow.
type AuthStrategy string

const (
	AuthStrategyAppInstallation AuthStrategy = "app_installation"
	AuthStrategyOAuthInstall    AuthStrategy = "oauth_install"
	AuthStrategyOAuthLink       AuthStrategy = "oauth_account_link"
	AuthStrategyWebhookOnly     AuthStrategy = "webhook_only"
)

// ConfigurationRequirement describes an environment/configuration field
// without containing its value. Sensitive values are redacted by config docs
// and must never have a default.
type ConfigurationRequirement struct {
	Key       string
	Required  bool
	Sensitive bool
	Purpose   string
}

// DisconnectPolicy makes provider cleanup behavior reviewable. Retention is
// for safe mappings/audit facts only, never raw credentials or webhook bodies.
type DisconnectPolicy struct {
	RevokeRemoteGrant      bool
	DeleteWebhook          bool
	DeleteCredentials      bool
	RetainMappingMetadata  bool
	MappingRetentionPeriod time.Duration
}

// Descriptor is immutable after registry construction. It contains discovery
// metadata only; typed adapter factories stay beside their consuming ports.
type Descriptor struct {
	Key             ProviderKey
	DisplayName     string
	Family          Family
	Capabilities    []Capability
	AuthStrategies  []AuthStrategy
	Configuration   []ConfigurationRequirement
	Disconnect      DisconnectPolicy
	OperatorRunbook string
}

func (descriptor Descriptor) clone() Descriptor {
	descriptor.Capabilities = append([]Capability(nil), descriptor.Capabilities...)
	descriptor.AuthStrategies = append([]AuthStrategy(nil), descriptor.AuthStrategies...)
	descriptor.Configuration = append([]ConfigurationRequirement(nil), descriptor.Configuration...)
	return descriptor
}
