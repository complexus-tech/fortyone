package documentshttp

import (
	"context"
	"net/http"
	"strings"
	"time"

	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const documentCommentAvatarAccessTTL = 24 * time.Hour

type documentCommentAvatarResolver interface {
	ResolveProfileImageURL(context.Context, string, time.Duration) (string, error)
}

func (h *Handlers) resolveCommentAvatar(ctx context.Context, comment *documents.CoreComment) {
	if comment == nil || comment.AuthorAvatar == nil || strings.TrimSpace(*comment.AuthorAvatar) == "" {
		return
	}
	resolver, ok := h.attachments.(documentCommentAvatarResolver)
	if !ok {
		comment.AuthorAvatar = nil
		return
	}
	resolved, err := resolver.ResolveProfileImageURL(ctx, strings.TrimSpace(*comment.AuthorAvatar), documentCommentAvatarAccessTTL)
	if err != nil || strings.TrimSpace(resolved) == "" {
		comment.AuthorAvatar = nil
		return
	}
	comment.AuthorAvatar = &resolved
}

func (h *Handlers) resolveCommentThreadAvatars(ctx context.Context, threads []documents.CoreCommentThread) {
	for threadIndex := range threads {
		for commentIndex := range threads[threadIndex].Comments {
			h.resolveCommentAvatar(ctx, &threads[threadIndex].Comments[commentIndex])
		}
	}
}

func (h *Handlers) ListComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	threads, err := h.documents.ListComments(ctx, workspace.ID, userID, documentID)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	h.resolveCommentThreadAvatars(ctx, threads)
	return web.Respond(ctx, w, toAppDocumentCommentThreads(threads), http.StatusOK)
}

func (h *Handlers) CreateComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	var input AppCreateDocumentComment
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	thread, err := h.documents.CreateComment(ctx, documents.CoreCreateCommentInput{
		WorkspaceID: workspace.ID, UserID: userID, DocumentID: documentID,
		Body: input.Body, Quote: input.Quote, AnchorStart: input.AnchorStart, AnchorEnd: input.AnchorEnd,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	for index := range thread.Comments {
		h.resolveCommentAvatar(ctx, &thread.Comments[index])
	}
	return web.Respond(ctx, w, toAppDocumentCommentThread(thread), http.StatusCreated)
}

func (h *Handlers) ReplyToComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	threadID, err := commentThreadIDFromRequest(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppDocumentCommentReply
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	comment, err := h.documents.ReplyToComment(ctx, documents.CoreReplyCommentInput{
		WorkspaceID: workspace.ID, UserID: userID, DocumentID: documentID,
		ThreadID: threadID, Body: input.Body,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	h.resolveCommentAvatar(ctx, &comment)
	return web.Respond(ctx, w, toAppDocumentComment(comment), http.StatusCreated)
}

func (h *Handlers) ResolveComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	threadID, err := commentThreadIDFromRequest(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppResolveDocumentComment
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.documents.ResolveComment(ctx, documents.CoreResolveCommentInput{
		WorkspaceID: workspace.ID, UserID: userID, DocumentID: documentID,
		ThreadID: threadID, Resolved: input.Resolved,
	}); err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func commentThreadIDFromRequest(r *http.Request) (uuid.UUID, error) {
	threadID, err := uuid.Parse(web.Params(r, "threadId"))
	if err != nil || threadID == uuid.Nil {
		return uuid.Nil, documents.ErrInvalidInput
	}
	return threadID, nil
}
