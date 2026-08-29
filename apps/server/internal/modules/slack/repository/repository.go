package slackrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrActiveInstallationConflict  = slackdomain.ErrConflict
	ErrWorkspaceAlreadyConnected   = slackdomain.ErrWorkspaceAlreadyConnected
	ErrSlackTeamAlreadyConnected   = slackdomain.ErrSlackTeamAlreadyConnected
	ErrUninstallInProgress         = slackdomain.ErrUninstallInProgress
	ErrUninstallResolutionRequired = slackdomain.ErrUninstallResolutionRequired
)

const (
	SlackUninstallMaxAttempts             = slackdomain.UninstallMaxAttempts
	slackUninstallLease                   = 2 * time.Minute
	SlackInstallationLifecycleAdvisoryKey = int64(0x534c41434b)
)

type Repo struct {
	queries        slacksql.Querier
	runTransaction func(context.Context, func(slacksql.Querier) error) error
}

// New composes Slack persistence from the native pgx pool. Logging and
// transport concerns deliberately stay outside this boundary.
func New(pool *pgxpool.Pool) *Repo {
	if pool == nil {
		return &Repo{}
	}
	queries := slacksql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.runTransaction = func(ctx context.Context, operation func(slacksql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{
			IsoLevel:   pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite,
		}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries slacksql.Querier) *Repo {
	repository := &Repo{queries: queries}
	if queries != nil {
		repository.runTransaction = func(ctx context.Context, operation func(slacksql.Querier) error) error {
			return operation(queries)
		}
	}
	return repository
}

func (repository *Repo) withinTransaction(
	ctx context.Context,
	operation func(slacksql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if repository == nil || repository.queries == nil || repository.runTransaction == nil {
		return errors.New("slack repository transactions are unavailable")
	}
	return mapDatabaseError(repository.runTransaction(ctx, operation))
}

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	for _, domainErr := range []error{
		slackdomain.ErrNotFound,
		slackdomain.ErrForbidden,
		slackdomain.ErrConflict,
		slackdomain.ErrInvalidInput,
		slackdomain.ErrWorkspaceAlreadyConnected,
		slackdomain.ErrSlackTeamAlreadyConnected,
		slackdomain.ErrUninstallInProgress,
		slackdomain.ErrUninstallResolutionRequired,
	} {
		if errors.Is(err, domainErr) {
			return err
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: persistence row does not exist", slackdomain.ErrNotFound)
	}
	switch platformdatabase.Classify(err) {
	case platformdatabase.ErrorClassSerializationFailure,
		platformdatabase.ErrorClassDeadlock,
		platformdatabase.ErrorClassUniqueViolation:
		return fmt.Errorf("%w: %v", slackdomain.ErrConflict, err)
	case platformdatabase.ErrorClassForeignKeyViolation,
		platformdatabase.ErrorClassNotNullViolation,
		platformdatabase.ErrorClassCheckViolation:
		return fmt.Errorf("%w: %v", slackdomain.ErrInvalidInput, err)
	default:
		return err
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, slackdomain.ErrNotFound) || errors.Is(err, pgx.ErrNoRows)
}

type (
	WorkspaceRecord       = slackdomain.Workspace
	TeamRecord            = slackdomain.Team
	StatusRecord          = slackdomain.Status
	TeamMemberRecord      = slackdomain.TeamMember
	LabelRecord           = slackdomain.Label
	ObjectiveRecord       = slackdomain.Objective
	WorkspaceMemberRecord = slackdomain.WorkspaceMember
	SlackUserLinkUpsert   = slackdomain.UserLinkUpsert
	SlackUserLinkRecord   = slackdomain.UserLink
	SlackWorkspaceRecord  = slackdomain.Installation
	SlackChannelRecord    = slackdomain.Channel
	OAuthInstallPayload   = slackdomain.OAuthInstallation
	SlackUninstallRecord  = slackdomain.Uninstall
	SlackUninstallInput   = slackdomain.EnqueueUninstall
	SlackChannelPayload   = slackdomain.ChannelUpsert
	SlackRequestLogInsert = slackdomain.RequestLogInsert
	SlackRequestLogRecord = slackdomain.RequestLog
)
