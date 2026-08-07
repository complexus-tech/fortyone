package documents

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	listed       CoreListInput
	created      CoreCreateInput
	duplicated   [3]uuid.UUID
	deleted      [3]uuid.UUID
	access       CoreAccessInput
	relationship CoreRelationshipInput
	media        CoreMediaInput
	mediaChecks  int
}

func (r *repositoryStub) List(_ context.Context, input CoreListInput) ([]CoreDocumentSummary, error) {
	r.listed = input
	return []CoreDocumentSummary{}, nil
}

func (r *repositoryStub) Get(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreDocument, error) {
	return CoreDocument{}, nil
}

func (r *repositoryStub) Create(_ context.Context, input CoreCreateInput) (CoreDocument, error) {
	r.created = input
	return CoreDocument{
		Title:       input.Title,
		Visibility:  input.Visibility,
		ContentHTML: input.ContentHTML,
		ContentText: input.ContentText,
	}, nil
}

func (r *repositoryStub) Duplicate(_ context.Context, workspaceID, userID, documentID uuid.UUID) (CoreDocument, error) {
	r.duplicated = [3]uuid.UUID{workspaceID, userID, documentID}
	return CoreDocument{ID: uuid.New(), Visibility: VisibilityPrivate}, nil
}

func (r *repositoryStub) Update(context.Context, CoreUpdateInput) (CoreDocument, error) {
	return CoreDocument{}, nil
}

func (r *repositoryStub) Archive(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}

func (r *repositoryStub) Delete(_ context.Context, workspaceID, userID, documentID uuid.UUID) ([]uuid.UUID, error) {
	r.deleted = [3]uuid.UUID{workspaceID, userID, documentID}
	return []uuid.UUID{uuid.New()}, nil
}

func (r *repositoryStub) SetAccess(_ context.Context, input CoreAccessInput) (CoreDocument, error) {
	r.access = input
	return CoreDocument{Visibility: input.Visibility, SharedWith: input.Members}, nil
}

func (r *repositoryStub) AddRelationship(_ context.Context, input CoreRelationshipInput) (CoreRelatedWork, error) {
	r.relationship = input
	return CoreRelatedWork{EntityID: input.EntityID, EntityType: input.EntityType}, nil
}

func (r *repositoryStub) RemoveRelationship(context.Context, CoreRelationshipInput) error {
	return nil
}

func (r *repositoryStub) ListRelatedDocuments(context.Context, uuid.UUID, uuid.UUID, RelationshipType, uuid.UUID) ([]CoreDocumentSummary, error) {
	return []CoreDocumentSummary{}, nil
}

func (r *repositoryStub) LinkMedia(_ context.Context, input CoreMediaInput) error {
	r.media = input
	return nil
}

func (r *repositoryStub) UnlinkMedia(_ context.Context, input CoreMediaInput) (bool, error) {
	r.media = input
	return true, nil
}

func (r *repositoryStub) AuthorizeMedia(_ context.Context, input CoreMediaInput) error {
	r.media = input
	r.mediaChecks++
	return nil
}

func TestCreateDefaultsBlankTitle(t *testing.T) {
	repo := &repositoryStub{}
	service := New(nil, repo)
	workspaceID, userID := uuid.New(), uuid.New()

	document, err := service.Create(context.Background(), CoreCreateInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Title:       "   ",
		Visibility:  VisibilityWorkspace,
	})

	require.NoError(t, err)
	require.Equal(t, "Untitled document", document.Title)
	require.Equal(t, VisibilityWorkspace, repo.created.Visibility)
}

func TestListLeavesLimitUnsetWhenNotRequested(t *testing.T) {
	repo := &repositoryStub{}
	service := New(nil, repo)

	_, err := service.List(context.Background(), CoreListInput{
		WorkspaceID: uuid.New(),
		UserID:      uuid.New(),
	})

	require.NoError(t, err)
	require.Nil(t, repo.listed.Limit)
}

func TestListCapsRequestedLimit(t *testing.T) {
	repo := &repositoryStub{}
	service := New(nil, repo)
	requestedLimit := maxListLimit + 1

	_, err := service.List(context.Background(), CoreListInput{
		WorkspaceID: uuid.New(),
		UserID:      uuid.New(),
		Limit:       &requestedLimit,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.listed.Limit)
	require.Equal(t, maxListLimit, *repo.listed.Limit)
}

func TestListRejectsNonPositiveLimit(t *testing.T) {
	service := New(nil, &repositoryStub{})

	for _, limit := range []int{0, -1} {
		limit := limit
		t.Run(fmt.Sprintf("limit_%d", limit), func(t *testing.T) {
			_, err := service.List(context.Background(), CoreListInput{
				WorkspaceID: uuid.New(),
				UserID:      uuid.New(),
				Limit:       &limit,
			})

			require.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

func TestCreateForwardsTemplateContent(t *testing.T) {
	repo := &repositoryStub{}
	service := New(nil, repo)
	contentHTML := "<h1>Meeting notes</h1><p>Agenda</p>"
	contentText := "Meeting notes\nAgenda"

	document, err := service.Create(context.Background(), CoreCreateInput{
		WorkspaceID: uuid.New(),
		UserID:      uuid.New(),
		Title:       "Meeting notes",
		Visibility:  VisibilityWorkspace,
		ContentHTML: contentHTML,
		ContentText: contentText,
	})

	require.NoError(t, err)
	require.Equal(t, contentHTML, repo.created.ContentHTML)
	require.Equal(t, contentText, repo.created.ContentText)
	require.Equal(t, contentHTML, document.ContentHTML)
	require.Equal(t, contentText, document.ContentText)
}

func TestDuplicateForwardsDocumentIdentity(t *testing.T) {
	repo := &repositoryStub{}
	service := New(nil, repo)
	workspaceID, userID, documentID := uuid.New(), uuid.New(), uuid.New()

	document, err := service.Duplicate(context.Background(), workspaceID, userID, documentID)

	require.NoError(t, err)
	require.Equal(t, [3]uuid.UUID{workspaceID, userID, documentID}, repo.duplicated)
	require.Equal(t, VisibilityPrivate, document.Visibility)
}

func TestDeleteRejectsIncompleteIdentity(t *testing.T) {
	repo := &repositoryStub{}
	service := New(nil, repo)

	attachments, err := service.Delete(context.Background(), uuid.New(), uuid.Nil, uuid.New())

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Nil(t, attachments)
	require.Equal(t, [3]uuid.UUID{}, repo.deleted)
}

func TestSetAccessDeduplicatesMembersAndExcludesOwner(t *testing.T) {
	repo := &repositoryStub{}
	service := New(nil, repo)
	ownerID, memberID := uuid.New(), uuid.New()

	_, err := service.SetAccess(context.Background(), CoreAccessInput{
		WorkspaceID: uuid.New(),
		UserID:      ownerID,
		DocumentID:  uuid.New(),
		Visibility:  VisibilityRestricted,
		Members: []CoreDocumentMember{
			{UserID: ownerID, Role: "editor"},
			{UserID: memberID, Role: "viewer"},
			{UserID: memberID, Role: "editor"},
		},
	})

	require.NoError(t, err)
	require.Equal(t, []CoreDocumentMember{{UserID: memberID, Role: "viewer"}}, repo.access.Members)
}

func TestRelationshipRejectsUnsupportedEntityType(t *testing.T) {
	service := New(nil, &repositoryStub{})

	_, err := service.AddRelationship(context.Background(), CoreRelationshipInput{
		WorkspaceID: uuid.New(),
		UserID:      uuid.New(),
		DocumentID:  uuid.New(),
		EntityType:  RelationshipType("sprint"),
		EntityID:    uuid.New(),
	})

	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestDocumentMediaRequiresCompleteIdentity(t *testing.T) {
	repo := &repositoryStub{}
	service := New(nil, repo)

	err := service.LinkMedia(context.Background(), CoreMediaInput{
		WorkspaceID: uuid.New(),
		UserID:      uuid.New(),
		DocumentID:  uuid.New(),
	})

	require.ErrorIs(t, err, ErrInvalidInput)
	require.Equal(t, CoreMediaInput{}, repo.media)
}

func TestAuthorizeDocumentMediaForwardsAccessIdentity(t *testing.T) {
	repo := &repositoryStub{}
	service := New(nil, repo)
	input := CoreMediaInput{
		WorkspaceID:  uuid.New(),
		UserID:       uuid.New(),
		DocumentID:   uuid.New(),
		AttachmentID: uuid.New(),
	}

	err := service.AuthorizeMedia(context.Background(), input)

	require.NoError(t, err)
	require.Equal(t, input, repo.media)
	require.Equal(t, 1, repo.mediaChecks)
}
