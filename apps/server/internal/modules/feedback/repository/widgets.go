package feedbackrepository

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type widgetSettingsProjection struct {
	PortalID                 uuid.UUID
	Enabled                  bool
	WidgetKeyID              uuid.UUID
	AllowedOrigins           []string
	SigningSecretEncrypted   *string
	SigningSecretVersion     int32
	PreviousVersionExpiresAt *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

func (projection widgetSettingsProjection) core() feedback.CoreWidgetSettings {
	secret := ""
	if projection.SigningSecretEncrypted != nil {
		secret = *projection.SigningSecretEncrypted
	}
	origins := append([]string(nil), projection.AllowedOrigins...)
	if origins == nil {
		origins = []string{}
	}
	return feedback.CoreWidgetSettings{PortalID: projection.PortalID, Enabled: projection.Enabled,
		WidgetKeyID: projection.WidgetKeyID, AllowedOrigins: origins, SigningSecretEncrypted: secret,
		SigningSecretVersion:     int(projection.SigningSecretVersion),
		PreviousVersionExpiresAt: projection.PreviousVersionExpiresAt,
		CreatedAt:                projection.CreatedAt, UpdatedAt: projection.UpdatedAt}
}

func (r *Repo) GetWidgetSettings(context.Context, uuid.UUID, uuid.UUID) (feedback.CoreWidgetSettings, error) {
	return feedback.CoreWidgetSettings{}, feedback.ErrForbidden
}

func (r *Repo) GetWidgetSettingsScoped(ctx context.Context, scope feedback.CoreAccessScope, portalID uuid.UUID) (feedback.CoreWidgetSettings, error) {
	if err := scope.Validate(); err != nil {
		return feedback.CoreWidgetSettings{}, feedback.ErrForbidden
	}
	if err := r.queries.EnsureFeedbackWidgetSettings(ctx, feedbacksql.EnsureFeedbackWidgetSettingsParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, PortalID: portalID,
	}); err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	row, err := r.queries.GetFeedbackWidgetSettings(ctx, feedbacksql.GetFeedbackWidgetSettingsParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, PortalID: portalID,
	})
	if err != nil {
		return feedback.CoreWidgetSettings{}, normalizeError(err)
	}
	return widgetSettingsProjection{row.PortalID, row.Enabled, row.WidgetKeyID, row.AllowedOrigins,
		row.SigningSecretEncrypted, row.SigningSecretVersion, timePointer(row.PreviousVersionExpiresAt),
		row.CreatedAt, row.UpdatedAt}.core(), nil
}

func (r *Repo) GetPublicWidgetSettings(ctx context.Context, portalID uuid.UUID) (feedback.CoreWidgetSettings, error) {
	row, err := r.queries.GetPublicFeedbackWidgetSettings(ctx, feedbacksql.GetPublicFeedbackWidgetSettingsParams{PortalID: portalID})
	if err != nil {
		return feedback.CoreWidgetSettings{}, normalizeError(err)
	}
	return widgetSettingsProjection{row.PortalID, row.Enabled, row.WidgetKeyID, row.AllowedOrigins,
		row.SigningSecretEncrypted, row.SigningSecretVersion, timePointer(row.PreviousVersionExpiresAt),
		row.CreatedAt, row.UpdatedAt}.core(), nil
}

func (r *Repo) UpsertWidgetSettings(context.Context, feedback.CoreWidgetSettingsInput) (feedback.CoreWidgetSettings, error) {
	return feedback.CoreWidgetSettings{}, feedback.ErrForbidden
}

func (r *Repo) UpsertWidgetSettingsScoped(ctx context.Context, scope feedback.CoreAccessScope, input feedback.CoreWidgetSettingsInput) (feedback.CoreWidgetSettings, error) {
	if err := scope.Validate(); err != nil || input.Access.WorkspaceID != scope.WorkspaceID ||
		input.Access.ActorID != scope.ActorID || input.WorkspaceID != scope.WorkspaceID {
		return feedback.CoreWidgetSettings{}, feedback.ErrForbidden
	}
	row, err := r.queries.UpsertFeedbackWidgetSettings(ctx, feedbacksql.UpsertFeedbackWidgetSettingsParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, PortalID: input.PortalID,
		Enabled: input.Enabled, AllowedOrigins: input.AllowedOrigins,
	})
	if err != nil {
		return feedback.CoreWidgetSettings{}, normalizeError(err)
	}
	return widgetSettingsProjection{row.PortalID, row.Enabled, row.WidgetKeyID, row.AllowedOrigins,
		row.SigningSecretEncrypted, row.SigningSecretVersion, timePointer(row.PreviousVersionExpiresAt),
		row.CreatedAt, row.UpdatedAt}.core(), nil
}

func (r *Repo) SetInitialWidgetSecret(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, int) (feedback.CoreWidgetSettings, error) {
	return feedback.CoreWidgetSettings{}, feedback.ErrForbidden
}

func (r *Repo) SetInitialWidgetSecretScoped(ctx context.Context, scope feedback.CoreAccessScope, portalID, keyID uuid.UUID, encrypted string, version int) (feedback.CoreWidgetSettings, error) {
	if err := scope.Validate(); err != nil {
		return feedback.CoreWidgetSettings{}, feedback.ErrForbidden
	}
	version32, err := safecast.Int32(version)
	if err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	row, err := r.queries.SetInitialFeedbackWidgetSecret(ctx, feedbacksql.SetInitialFeedbackWidgetSecretParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, PortalID: portalID,
		WidgetKeyID: keyID, SigningSecretEncrypted: nonEmptyStringPointer(encrypted), SigningSecretVersion: version32,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return feedback.CoreWidgetSettings{}, fmt.Errorf("%w: widget signing secret already exists", feedback.ErrInvalidInput)
	}
	if err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	return widgetSettingsProjection{row.PortalID, row.Enabled, row.WidgetKeyID, row.AllowedOrigins,
		row.SigningSecretEncrypted, row.SigningSecretVersion, row.PreviousVersionExpiresAt,
		row.CreatedAt, row.UpdatedAt}.core(), nil
}

func (r *Repo) RotateWidgetSecret(context.Context, uuid.UUID, uuid.UUID, string, int, time.Time) (feedback.CoreWidgetSettings, error) {
	return feedback.CoreWidgetSettings{}, feedback.ErrForbidden
}

func (r *Repo) RotateWidgetSecretScoped(ctx context.Context, scope feedback.CoreAccessScope, portalID uuid.UUID, encrypted string, version int, graceExpiresAt time.Time) (feedback.CoreWidgetSettings, error) {
	if err := scope.Validate(); err != nil {
		return feedback.CoreWidgetSettings{}, feedback.ErrForbidden
	}
	version32, err := safecast.Int32(version)
	if err != nil {
		return feedback.CoreWidgetSettings{}, err
	}
	var settings feedback.CoreWidgetSettings
	err = r.withinTransaction(ctx, pgx.TxOptions{}, func(q feedbacksql.Querier) error {
		current, err := q.LockFeedbackWidgetSettings(ctx, feedbacksql.LockFeedbackWidgetSettingsParams{
			ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, PortalID: portalID,
		})
		if err != nil {
			return normalizeError(err)
		}
		if current.SigningSecretEncrypted == nil || current.SigningSecretVersion <= 0 || version32 != current.SigningSecretVersion+1 {
			return feedback.ErrInvalidInput
		}
		if err := q.SavePreviousFeedbackWidgetSecret(ctx, feedbacksql.SavePreviousFeedbackWidgetSecretParams{
			PortalID: portalID, SigningSecretVersion: current.SigningSecretVersion,
			SigningSecretEncrypted: *current.SigningSecretEncrypted, ActivatedAt: current.UpdatedAt,
			GraceExpiresAt: graceExpiresAt,
		}); err != nil {
			return err
		}
		count, err := q.UpdateFeedbackWidgetSecret(ctx, feedbacksql.UpdateFeedbackWidgetSecretParams{
			SigningSecretEncrypted: nonEmptyStringPointer(encrypted), SigningSecretVersion: version32,
			PortalID: portalID, PreviousVersion: current.SigningSecretVersion,
		})
		if err != nil {
			return err
		}
		if err := requireRowsAffected(count); err != nil {
			return feedback.ErrVersionConflict
		}
		row, err := q.GetFeedbackWidgetSettings(ctx, feedbacksql.GetFeedbackWidgetSettingsParams{
			ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, PortalID: portalID,
		})
		if err != nil {
			return normalizeError(err)
		}
		settings = widgetSettingsProjection{row.PortalID, row.Enabled, row.WidgetKeyID, row.AllowedOrigins,
			row.SigningSecretEncrypted, row.SigningSecretVersion, timePointer(row.PreviousVersionExpiresAt),
			row.CreatedAt, row.UpdatedAt}.core()
		return nil
	})
	return settings, err
}

func (r *Repo) GetWidgetSigningSecret(ctx context.Context, portalID, widgetKeyID uuid.UUID, version int) (string, error) {
	version32, err := safecast.Int32(version)
	if err != nil {
		return "", err
	}
	current, currentErr := r.queries.GetCurrentFeedbackWidgetSigningSecret(ctx, feedbacksql.GetCurrentFeedbackWidgetSigningSecretParams{
		PortalID: portalID, WidgetKeyID: widgetKeyID, SigningSecretVersion: version32,
	})
	if currentErr == nil && current != nil {
		return *current, nil
	}
	if currentErr != nil && !errors.Is(currentErr, pgx.ErrNoRows) {
		return "", currentErr
	}
	previous, err := r.queries.GetPreviousFeedbackWidgetSigningSecret(ctx, feedbacksql.GetPreviousFeedbackWidgetSigningSecretParams{
		PortalID: portalID, WidgetKeyID: widgetKeyID, SigningSecretVersion: version32,
	})
	return previous, normalizeError(err)
}

func (r *Repo) ConsumeWidgetAssertionNonce(ctx context.Context, portalID, widgetKeyID uuid.UUID, version int, nonce, parentOrigin string, expiresAt time.Time) error {
	version32, err := safecast.Int32(version)
	if err != nil {
		return err
	}
	nonceHash := sha256.Sum256([]byte(nonce))
	err = r.queries.ConsumeFeedbackWidgetAssertionNonce(ctx, feedbacksql.ConsumeFeedbackWidgetAssertionNonceParams{
		SigningSecretVersion: version32, NonceHash: nonceHash[:], ParentOrigin: parentOrigin,
		ExpiresAt: expiresAt, PortalID: portalID, WidgetKeyID: widgetKeyID,
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return feedback.ErrWidgetAssertionReplayed
	}
	return normalizeError(err)
}
