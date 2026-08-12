package emailagent

import "strings"

// ParseControlCommand recognizes only a complete current-reply body containing
// CONFIRM or CANCEL, case-insensitively, after surrounding whitespace is
// removed. The inbound adapter must strip quoted history and signatures first.
// Natural affirmations such as "yes" deliberately remain model input.
func ParseControlCommand(currentReply string) (ControlKind, bool) {
	normalized := strings.TrimSpace(currentReply)
	switch {
	case strings.EqualFold(normalized, "CONFIRM"):
		return ControlConfirm, true
	case strings.EqualFold(normalized, "CANCEL"):
		return ControlCancel, true
	default:
		return "", false
	}
}
