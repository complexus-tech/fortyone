// Package slackdomain contains transport- and persistence-neutral Slack
// integration state. Provider payloads and generated database rows must not
// cross this boundary.
package slackdomain

import "errors"

var (
	ErrNotFound                    = errors.New("slack resource not found")
	ErrForbidden                   = errors.New("slack integration access is forbidden")
	ErrConflict                    = errors.New("slack integration state conflicts with another operation")
	ErrInvalidInput                = errors.New("invalid Slack integration input")
	ErrWorkspaceAlreadyConnected   = errors.New("this FortyOne workspace is already connected to another Slack team")
	ErrSlackTeamAlreadyConnected   = errors.New("this Slack team is already connected to another FortyOne workspace")
	ErrUninstallInProgress         = errors.New("slack uninstall is still processing")
	ErrUninstallResolutionRequired = errors.New("slack uninstall requires operator resolution")
)
