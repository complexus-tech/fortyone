package credentialvault

const (
	// DefaultMaintenanceBatchSize bounds each provider credential query while
	// keeping startup/operator maintenance efficient for ordinary installations.
	DefaultMaintenanceBatchSize = 100
	// MaxMaintenanceBatchSize prevents an operator-controlled batch from causing
	// an unbounded allocation or long-lived database statement.
	MaxMaintenanceBatchSize = 500
)

// RotationReport contains only safe operational metadata. It intentionally
// excludes credential envelopes, provider account names, and plaintext.
type RotationReport struct {
	ActiveKey KeyRef
	Scanned   int
	Current   int
	Rewrapped int
	Stale     int
}

// Add combines reports produced by independently paged provider stores.
func (report *RotationReport) Add(other RotationReport) {
	if report == nil {
		return
	}
	if report.ActiveKey == (KeyRef{}) {
		report.ActiveKey = other.ActiveKey
	}
	report.Scanned += other.Scanned
	report.Current += other.Current
	report.Rewrapped += other.Rewrapped
	report.Stale += other.Stale
}

// MaintenanceBatchSize applies the shared bounded-query policy.
func MaintenanceBatchSize(requested int) int {
	if requested <= 0 {
		return DefaultMaintenanceBatchSize
	}
	if requested > MaxMaintenanceBatchSize {
		return MaxMaintenanceBatchSize
	}
	return requested
}
