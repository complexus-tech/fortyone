package feedbackhttp

import (
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
)

type AppVerificationRequest struct {
	Email            string `json:"email"`
	DisplayName      string `json:"displayName"`
	HideNamePublicly bool   `json:"hideNamePublicly"`
	Source           string `json:"source"`
}

type AppVerificationAccepted struct {
	Accepted  bool       `json:"accepted"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type AppVerificationConfirm struct {
	Token  string `json:"token"`
	Email  string `json:"email"`
	Code   string `json:"code"`
	Source string `json:"source"`
}

type AppParticipant struct {
	ID          uuid.UUID `json:"id"`
	Kind        string    `json:"kind"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"displayName"`
	PublicName  string    `json:"publicName"`
	AvatarURL   *string   `json:"avatarUrl"`
	Masked      bool      `json:"masked"`
}

type AppContributorSession struct {
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type AppContributorSessionResponse struct {
	Participant       AppParticipant        `json:"participant"`
	Session           AppContributorSession `json:"session"`
	UnreadUpdateCount int                   `json:"unreadUpdateCount"`
}

type AppFollowState struct {
	ItemID    uuid.UUID `json:"itemId"`
	Following bool      `json:"following"`
}

type AppPreferenceExchange struct {
	Token string `json:"token"`
}

type AppPreferenceItem struct {
	ItemID    uuid.UUID `json:"itemId"`
	ItemSlug  string    `json:"itemSlug"`
	Title     string    `json:"title"`
	Following bool      `json:"following"`
}

type AppContributorPreferences struct {
	PortalEmailsEnabled bool                `json:"portalEmailsEnabled"`
	Items               []AppPreferenceItem `json:"items"`
}

type AppUpdatePreferences struct {
	PortalEmailsEnabled *bool               `json:"portalEmailsEnabled"`
	Items               []AppPreferenceItem `json:"items"`
}

type AppUpdateItem struct {
	ID     uuid.UUID `json:"id"`
	Slug   string    `json:"slug"`
	Title  string    `json:"title"`
	Status string    `json:"status"`
}

type AppFeedbackUpdate struct {
	ID                uuid.UUID       `json:"id"`
	WorkspaceID       uuid.UUID       `json:"workspaceId"`
	PortalID          uuid.UUID       `json:"portalId"`
	Slug              string          `json:"slug"`
	Title             string          `json:"title"`
	Summary           *string         `json:"summary"`
	Body              string          `json:"body"`
	CoverImageURL     *string         `json:"coverImageUrl"`
	Status            string          `json:"status"`
	PublishedAt       *time.Time      `json:"publishedAt"`
	PublishedByUserID *uuid.UUID      `json:"publishedByUserId"`
	LinkedItems       []AppUpdateItem `json:"linkedItems"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

type AppFeedbackUpdatesPage struct {
	Updates     []AppFeedbackUpdate `json:"updates"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"pageSize"`
	HasMore     bool                `json:"hasMore"`
	UnreadCount int                 `json:"unreadCount,omitempty"`
}

type AppUpdateInput struct {
	PortalID      uuid.UUID   `json:"portalId"`
	Title         string      `json:"title"`
	Slug          string      `json:"slug"`
	Summary       *string     `json:"summary"`
	Body          string      `json:"body"`
	CoverImageURL *string     `json:"coverImageUrl"`
	ItemIDs       []uuid.UUID `json:"itemIds"`
}

type AppUpdatesSeen struct {
	UnreadUpdateCount int       `json:"unreadUpdateCount"`
	LastSeenAt        time.Time `json:"lastSeenAt"`
}

type AppWidgetSettingsInput struct {
	Enabled        bool     `json:"enabled"`
	AllowedOrigins []string `json:"allowedOrigins"`
}

type AppWidgetSettings struct {
	PortalID                 uuid.UUID  `json:"portalId"`
	Enabled                  bool       `json:"enabled"`
	WidgetKeyID              uuid.UUID  `json:"widgetKeyId"`
	AllowedOrigins           []string   `json:"allowedOrigins"`
	SigningSecretVersion     int        `json:"signingSecretVersion"`
	HasSigningSecret         bool       `json:"hasSigningSecret"`
	PreviousVersionExpiresAt *time.Time `json:"previousVersionExpiresAt"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                time.Time  `json:"updatedAt"`
	SigningSecret            string     `json:"signingSecret,omitempty"`
}

type AppPublicWidgetConfig struct {
	Enabled        bool      `json:"enabled"`
	WidgetKeyID    uuid.UUID `json:"widgetKeyId"`
	AllowedOrigins []string  `json:"allowedOrigins"`
}

type AppWidgetSessionRequest struct {
	Assertion    string `json:"assertion"`
	ParentOrigin string `json:"parentOrigin"`
}

func toAppParticipant(participant feedback.CoreParticipant) AppParticipant {
	publicName := participant.DisplayName
	if publicName == "" {
		publicName = "Guest"
	}
	masked := participant.PublicMasked || participant.Kind == feedback.ContributorKindAnonymous
	if masked {
		publicName = "Anonymous"
	}
	return AppParticipant{
		ID: participant.ID, Kind: participant.Kind, Email: participant.Email,
		DisplayName: participant.DisplayName, PublicName: publicName, AvatarURL: participant.AvatarURL, Masked: masked,
	}
}

func toAppSession(result feedback.CoreContributorSessionResult, includeToken bool, unreadCount int) AppContributorSessionResponse {
	token := ""
	if includeToken {
		token = result.Token
	}
	return AppContributorSessionResponse{
		Participant:       toAppParticipant(result.Participant),
		Session:           AppContributorSession{Token: token, ExpiresAt: result.Session.ExpiresAt},
		UnreadUpdateCount: unreadCount,
	}
}

func toAppPreferences(preferences feedback.CoreContributorPreferences) AppContributorPreferences {
	items := make([]AppPreferenceItem, 0, len(preferences.ItemFollows))
	for _, item := range preferences.ItemFollows {
		items = append(items, AppPreferenceItem{ItemID: item.ItemID, ItemSlug: item.ItemSlug, Title: item.Title, Following: item.Following})
	}
	return AppContributorPreferences{PortalEmailsEnabled: preferences.PortalEmailsEnabled, Items: items}
}

func toAppUpdate(update feedback.CoreFeedbackUpdate) AppFeedbackUpdate {
	items := make([]AppUpdateItem, 0, len(update.LinkedItems))
	for _, item := range update.LinkedItems {
		items = append(items, AppUpdateItem{ID: item.ID, Slug: item.Slug, Title: item.Title, Status: item.Status})
	}
	return AppFeedbackUpdate{
		ID: update.ID, WorkspaceID: update.WorkspaceID, PortalID: update.PortalID, Slug: update.Slug,
		Title: update.Title, Summary: update.Summary, Body: update.Body, CoverImageURL: update.CoverImageURL,
		Status: update.Status, PublishedAt: update.PublishedAt, PublishedByUserID: update.PublishedByUserID,
		LinkedItems: items, CreatedAt: update.CreatedAt, UpdatedAt: update.UpdatedAt,
	}
}

func toAppUpdatesPage(page feedback.CoreUpdatesPage, unreadCount int) AppFeedbackUpdatesPage {
	updates := make([]AppFeedbackUpdate, 0, len(page.Updates))
	for _, update := range page.Updates {
		updates = append(updates, toAppUpdate(update))
	}
	return AppFeedbackUpdatesPage{Updates: updates, Page: page.Page, PageSize: page.PageSize, HasMore: page.HasMore, UnreadCount: unreadCount}
}

func toAppWidgetSettings(settings feedback.CoreWidgetSettings, signingSecret string) AppWidgetSettings {
	return AppWidgetSettings{
		PortalID: settings.PortalID, Enabled: settings.Enabled, WidgetKeyID: settings.WidgetKeyID,
		AllowedOrigins: settings.AllowedOrigins, SigningSecretVersion: settings.SigningSecretVersion,
		HasSigningSecret: settings.SigningSecretEncrypted != "", PreviousVersionExpiresAt: settings.PreviousVersionExpiresAt,
		CreatedAt: settings.CreatedAt, UpdatedAt: settings.UpdatedAt, SigningSecret: signingSecret,
	}
}
