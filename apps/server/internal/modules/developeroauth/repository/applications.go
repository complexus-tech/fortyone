package developeroauthrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauthsql "github.com/complexus-tech/projects-api/internal/modules/developeroauth/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

func (store *Store) CreateApplication(
	ctx context.Context,
	command developeroauthdomain.RegisterApplication,
) (developeroauthdomain.Application, error) {
	tx, queries, err := store.begin(ctx)
	if err != nil {
		return developeroauthdomain.Application{}, err
	}
	defer rollback(ctx, tx)
	row, err := queries.CreateOAuthApplication(ctx, developeroauthsql.CreateOAuthApplicationParams{
		ApplicationID: command.ID, ClientID: command.ClientID, Name: command.Name,
		RegistrationKind: command.RegistrationKind, ExpiresAt: command.ExpiresAt, CreatedAt: command.CreatedAt,
	})
	if err != nil {
		return developeroauthdomain.Application{}, fmt.Errorf("create OAuth application: %w", err)
	}
	for _, redirectURI := range command.RedirectURIs {
		if err := queries.CreateOAuthApplicationRedirectURI(ctx, developeroauthsql.CreateOAuthApplicationRedirectURIParams{
			ApplicationID: row.ApplicationID, RedirectURI: redirectURI, CreatedAt: command.CreatedAt,
		}); err != nil {
			return developeroauthdomain.Application{}, fmt.Errorf("create OAuth application redirect URI: %w", err)
		}
	}
	if err := createAuditEvent(ctx, queries, developeroauthdomain.AuditEvent{
		ID: uuidForAudit(), ApplicationID: &row.ApplicationID, Operation: "application.registered",
		Result: "succeeded", Metadata: []byte(`{"registration_kind":"dynamic_public"}`), CreatedAt: command.CreatedAt,
	}); err != nil {
		return developeroauthdomain.Application{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return developeroauthdomain.Application{}, fmt.Errorf("commit OAuth application: %w", err)
	}
	return developeroauthdomain.Application{
		ID: row.ApplicationID, ClientID: row.ClientID, Name: row.Name,
		RegistrationKind: row.RegistrationKind, RedirectURIs: append([]string(nil), command.RedirectURIs...),
		ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
	}, nil
}

func (store *Store) GetActiveApplication(
	ctx context.Context,
	clientID string,
	activeAt time.Time,
) (developeroauthdomain.Application, error) {
	row, err := store.queries.GetActiveOAuthApplicationByClientID(ctx, developeroauthsql.GetActiveOAuthApplicationByClientIDParams{
		ClientID: clientID, ActiveAt: activeAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return developeroauthdomain.Application{}, developeroauthdomain.ErrApplicationNotFound
	}
	if err != nil {
		return developeroauthdomain.Application{}, fmt.Errorf("get OAuth application: %w", err)
	}
	return developeroauthdomain.Application{
		ID: row.ApplicationID, ClientID: row.ClientID, Name: row.Name,
		RegistrationKind: row.RegistrationKind, RedirectURIs: append([]string(nil), row.RedirectUris...),
		ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
	}, nil
}
