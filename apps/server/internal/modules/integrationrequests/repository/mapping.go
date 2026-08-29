package integrationrequestsrepository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	integrationrequestssql "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func integrationRequestFromSQL(row integrationrequestssql.IntegrationRequest) (integrationrequestdomain.IntegrationRequest, error) {
	metadata := make(map[string]any)
	if len(row.Metadata) > 0 {
		if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
			return integrationrequestdomain.IntegrationRequest{}, fmt.Errorf("decode integration request metadata: %w", err)
		}
		if metadata == nil {
			metadata = make(map[string]any)
		}
	}
	return integrationrequestdomain.IntegrationRequest{
		ID: row.ID, WorkspaceID: row.WorkspaceID, TeamID: row.TeamID,
		Provider: row.Provider, SourceType: row.SourceType, SourceExternalID: row.SourceExternalID,
		SourceNumber: intFromInt32(row.SourceNumber), SourceURL: row.SourceURL,
		Title: row.Title, Description: row.Description, StatusID: row.StatusID, Priority: row.Priority,
		AssigneeID: row.AssigneeID, EstimateValue: row.EstimateUnit,
		EstimatedDurationMinutes: intFromInt32(row.EstimatedDurationMinutes),
		MinimumFocusBlockMinutes: intFromInt32(row.MinimumFocusBlockMinutes),
		ObjectiveID:              row.ObjectiveID, KeyResultID: row.KeyResultID, SprintID: row.SprintID,
		StartDate: row.StartDate, EndDate: row.EndDate, LabelIDs: append([]uuid.UUID(nil), row.LabelIds...),
		Status: row.Status, Metadata: metadata, AcceptedStoryID: row.AcceptedStoryID,
		AcceptedByUserID: row.AcceptedByUserID, AcceptedAt: row.AcceptedAt,
		DeclinedByUserID: row.DeclinedByUserID, DeclinedAt: row.DeclinedAt,
		AcceptanceState:           row.AcceptanceState,
		AcceptanceStartedByUserID: row.AcceptanceStartedByUserID,
		AcceptanceStartedAt:       row.AcceptanceStartedAt, CreatedByUserID: row.CreatedByUserID,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

type providerThreadRecord struct {
	ID                      uuid.UUID
	WorkspaceID             uuid.UUID
	IntegrationRequestID    uuid.UUID
	TeamID                  uuid.UUID
	AcceptedStoryID         *uuid.UUID
	Provider                string
	ExternalWorkspaceID     string
	InstallationGeneration  *uuid.UUID
	ExternalChannelID       string
	ExternalThreadID        string
	ExternalSourceMessageID *string
	SourceURL               *string
	RequestTitle            string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func providerThreadFromRecord(row providerThreadRecord) integrationrequestdomain.ProviderThread {
	return integrationrequestdomain.ProviderThread{
		ID: row.ID, WorkspaceID: row.WorkspaceID, IntegrationRequestID: row.IntegrationRequestID,
		TeamID: row.TeamID, AcceptedStoryID: row.AcceptedStoryID, Provider: row.Provider,
		ExternalWorkspaceID: row.ExternalWorkspaceID, InstallationGeneration: row.InstallationGeneration,
		ExternalChannelID: row.ExternalChannelID, ExternalThreadID: row.ExternalThreadID,
		ExternalSourceMessageID: row.ExternalSourceMessageID, SourceURL: row.SourceURL,
		RequestTitle: row.RequestTitle, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

type commentRecord struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	ThreadID               uuid.UUID
	Direction              string
	AuthorUserID           *uuid.UUID
	AuthorName             string
	AuthorAvatar           *string
	ExternalAuthorID       *string
	ExternalMessageID      *string
	ClientIdempotencyKey   *uuid.UUID
	OutboundIdempotencyKey *string
	DeliveryStatus         *string
	Body                   string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func commentFromRecord(row commentRecord) integrationrequestdomain.Comment {
	return integrationrequestdomain.Comment{
		ID: row.ID, WorkspaceID: row.WorkspaceID, ThreadID: row.ThreadID, Direction: row.Direction,
		AuthorUserID: row.AuthorUserID, AuthorName: row.AuthorName, AuthorAvatar: row.AuthorAvatar,
		ExternalAuthorID: row.ExternalAuthorID, ExternalMessageID: row.ExternalMessageID,
		ClientIdempotencyKey: row.ClientIdempotencyKey, OutboundIdempotencyKey: row.OutboundIdempotencyKey,
		DeliveryStatus: row.DeliveryStatus, Body: row.Body, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func mapNotFound(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", operation, integrationrequestdomain.ErrNotFound)
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func mapProviderThreadNotFound(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", operation, integrationrequestdomain.ErrProviderThreadNotFound)
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func intFromInt32(value *int32) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

func optionalValue[T any](current *T, patch integrationrequestdomain.OptionalValue[T]) *T {
	if patch.Set {
		return patch.Value
	}
	return current
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func uuidPointer(value uuid.UUID) *uuid.UUID { return &value }
func stringPointer(value string) *string     { return &value }
