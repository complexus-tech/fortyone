package chatservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
)

type Repository interface {
	WithinUnit(context.Context, func(context.Context) error) error
	ListConversations(context.Context, domain.ListQuery, uuid.UUID) (domain.ConversationPage, error)
	Unread(context.Context, domain.ActorScope) (domain.UnreadSummary, error)
	UnreadThreads(context.Context, domain.UnreadThreadQuery) (domain.UnreadThreadPage, error)
	UnreadThreadReadTargets(context.Context, domain.ActorScope) ([]domain.ThreadReadTarget, error)
	GetConversation(context.Context, domain.ActorScope, uuid.UUID, uuid.UUID) (domain.Conversation, error)
	CreateDirect(context.Context, domain.ActorScope, uuid.UUID, uuid.UUID) (domain.Conversation, error)
	ListMessages(context.Context, domain.MessageQuery) (domain.MessagePage, error)
	GetMessage(context.Context, domain.ActorScope, uuid.UUID) (domain.Message, error)
	CreateMessage(context.Context, domain.CreateMessageCommand, uuid.UUID) (domain.Message, error)
	UpdateMessage(context.Context, domain.UpdateMessageCommand) (domain.Message, error)
	DeleteMessage(context.Context, domain.DeleteMessageCommand) (domain.DeletedMessage, error)
	SetReaction(context.Context, domain.ActorScope, uuid.UUID, string, bool) (domain.Message, error)
	MarkRead(context.Context, domain.ReadCommand) (domain.ReadState, error)
	MarkUnread(context.Context, domain.MarkUnreadCommand) (domain.MarkUnreadState, error)
	Search(context.Context, domain.SearchQuery) (domain.SearchPage, error)
	Resources(context.Context, domain.SearchQuery) (domain.ResourcePage, error)
	BoundedContext(context.Context, domain.ContextQuery) ([]domain.Message, error)
	LinkPendingAttachment(context.Context, domain.AttachmentScope, uuid.UUID, domain.AttachmentMetadata) error
	AuthorizeAttachment(context.Context, domain.AttachmentScope, uuid.UUID) error
	UnlinkPendingAttachment(context.Context, domain.AttachmentScope, uuid.UUID) error
	AuthorizeTyping(context.Context, domain.TypingCommand) (domain.Person, error)
	TypingRecipientIDs(context.Context, domain.TypingCommand) ([]uuid.UUID, error)
}

type Service struct {
	rollout     *chatrollout.Policy
	repo        Repository
	mayaID      uuid.UUID
	created     domain.MessageCreatedHook
	afterCommit func(context.Context)
	typing      TypingBroadcaster
}

type Option func(*Service)

// WithWorkspaceSettings applies the live workspace setting before all service reads
// and writes, including calls made by Maya rather than an HTTP handler.
func WithWorkspaceSettings(policy chatrollout.Policy) Option {
	return func(s *Service) { s.rollout = &policy }
}

func (s *Service) validateScope(ctx context.Context, scope domain.ActorScope, write bool) error {
	if s.rollout != nil {
		enabled, err := s.rollout.Enabled(ctx, scope.WorkspaceID)
		if err != nil {
			return err
		}
		if !enabled {
			return domain.ErrForbidden
		}
	}
	return validateScope(scope, write)
}

func (s *Service) validateAttachmentScope(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID, write bool) error {
	if err := s.validateScope(ctx, scope.ActorScope, write); err != nil {
		return err
	}
	return validateAttachmentScope(scope, id, write)
}

func WithMayaActorID(id uuid.UUID) Option { return func(s *Service) { s.mayaID = id } }

func WithMessageCreatedHook(hook domain.MessageCreatedHook) Option {
	return func(s *Service) { s.created = hook }
}

func WithAfterCommit(callback func(context.Context)) Option {
	return func(s *Service) { s.afterCommit = callback }
}

func New(repo Repository, options ...Option) *Service {
	s := &Service{repo: repo}
	for _, option := range options {
		option(s)
	}
	return s
}

func validateScope(scope domain.ActorScope, write bool) error {
	permission := platformauth.ScopeChatRead
	if write {
		permission = platformauth.ScopeChatWrite
	}
	if scope.WorkspaceID == uuid.Nil || scope.Actor.Validate() != nil || !scope.Actor.IsUserActor() || scope.Actor.WorkspaceID != scope.WorkspaceID || !scope.Actor.Scopes.Has(permission) {
		return domain.ErrForbidden
	}
	return nil
}

func boundedLimit(limit, fallback, maximum int) (int, error) {
	if limit == 0 {
		return fallback, nil
	}
	if limit < 1 || limit > maximum {
		return 0, domain.ErrInvalid
	}
	return limit, nil
}

func (s *Service) mutate(ctx context.Context, operation func(context.Context) error) error {
	return s.repo.WithinUnit(ctx, func(unit context.Context) error {
		if err := operation(unit); err != nil {
			return err
		}
		if s.afterCommit != nil {
			platformdatabase.AfterCommit(unit, func() { s.afterCommit(context.WithoutCancel(ctx)) })
		}
		return nil
	})
}

func (s *Service) ListConversations(ctx context.Context, q domain.ListQuery) (domain.ConversationPage, error) {
	if err := s.validateScope(ctx, q.Scope, false); err != nil {
		return domain.ConversationPage{}, err
	}
	var err error
	if utf8.RuneCountInString(q.Search) > 200 {
		return domain.ConversationPage{}, domain.ErrInvalid
	}
	q.Limit, err = boundedLimit(q.Limit, 100, 200)
	if err != nil {
		return domain.ConversationPage{}, err
	}
	return s.repo.ListConversations(ctx, q, s.mayaID)
}

func (s *Service) GetConversation(ctx context.Context, scope domain.ActorScope, id uuid.UUID) (domain.Conversation, error) {
	if err := s.validateScope(ctx, scope, false); err != nil {
		return domain.Conversation{}, err
	}
	if id == uuid.Nil {
		return domain.Conversation{}, domain.ErrInvalid
	}
	return s.repo.GetConversation(ctx, scope, id, s.mayaID)
}

func (s *Service) CreateDirect(ctx context.Context, scope domain.ActorScope, peer uuid.UUID) (result domain.Conversation, err error) {
	if err = s.validateScope(ctx, scope, true); err != nil {
		return
	}
	if !scope.Actor.TeamAccess.IsUnrestricted() {
		return result, domain.ErrForbidden
	}
	if peer == uuid.Nil || peer == scope.Actor.PrincipalID {
		return result, domain.ErrInvalid
	}
	err = s.mutate(ctx, func(unit context.Context) error {
		result, err = s.repo.CreateDirect(unit, scope, peer, s.mayaID)
		return err
	})
	return
}

func (s *Service) ListMessages(ctx context.Context, q domain.MessageQuery) (domain.MessagePage, error) {
	if err := s.validateScope(ctx, q.Scope, false); err != nil {
		return domain.MessagePage{}, err
	}
	var err error
	q.Limit, err = boundedLimit(q.Limit, 50, 100)
	if err != nil {
		return domain.MessagePage{}, err
	}
	if q.ConversationID == uuid.Nil || (q.Before != "" && q.After != "") || (q.Around != nil && (q.Before != "" || q.After != "")) {
		return domain.MessagePage{}, domain.ErrInvalid
	}
	return s.repo.ListMessages(ctx, q)
}

func (s *Service) GetMessage(ctx context.Context, scope domain.ActorScope, id uuid.UUID) (domain.Message, error) {
	if err := s.validateScope(ctx, scope, false); err != nil {
		return domain.Message{}, err
	}
	if id == uuid.Nil {
		return domain.Message{}, domain.ErrInvalid
	}
	return s.repo.GetMessage(ctx, scope, id)
}

func (s *Service) CreateMessage(ctx context.Context, c domain.CreateMessageCommand) (result domain.Message, err error) {
	if err = s.validateScope(ctx, c.Scope, true); err != nil {
		return
	}
	if c.ConversationID == uuid.Nil || c.CreationID == uuid.Nil {
		return result, domain.ErrInvalid
	}
	if err = validateContent(c.Text, c.RichContent, c.MentionedUserIDs, c.AgentMentions, c.AttachmentIDs); err != nil {
		return
	}
	err = s.mutate(ctx, func(unit context.Context) error {
		var e error
		result, e = s.repo.CreateMessage(unit, c, c.Scope.Actor.PrincipalID)
		if e != nil {
			return e
		}
		if !result.Replayed && s.created != nil {
			return s.created(unit, c, result)
		}
		return nil
	})
	return
}

func (s *Service) UpdateMessage(ctx context.Context, c domain.UpdateMessageCommand) (result domain.Message, err error) {
	if err = s.validateScope(ctx, c.Scope, true); err != nil {
		return
	}
	if c.MessageID == uuid.Nil || c.Version < 1 {
		return result, domain.ErrInvalid
	}
	if err = validateContent(c.Text, c.RichContent, c.MentionedUserIDs, c.AgentMentions, c.AttachmentIDs); err != nil {
		return
	}
	err = s.mutate(ctx, func(unit context.Context) error { result, err = s.repo.UpdateMessage(unit, c); return err })
	return
}

func (s *Service) DeleteMessage(ctx context.Context, c domain.DeleteMessageCommand) (result domain.DeletedMessage, err error) {
	if err = s.validateScope(ctx, c.Scope, true); err != nil {
		return
	}
	if c.MessageID == uuid.Nil || c.Version < 1 {
		return result, domain.ErrInvalid
	}
	err = s.mutate(ctx, func(unit context.Context) error { result, err = s.repo.DeleteMessage(unit, c); return err })
	return
}

func (s *Service) SetReaction(ctx context.Context, scope domain.ActorScope, id uuid.UUID, emoji string, active bool) (result domain.Message, err error) {
	if err = s.validateScope(ctx, scope, true); err != nil {
		return
	}
	if id == uuid.Nil || !commentsdomain.SupportsReaction(emoji) {
		return result, domain.ErrInvalid
	}
	err = s.mutate(ctx, func(unit context.Context) error {
		result, err = s.repo.SetReaction(unit, scope, id, emoji, active)
		return err
	})
	return
}

func (s *Service) MarkRead(ctx context.Context, c domain.ReadCommand) (result domain.ReadState, err error) {
	if err = s.validateScope(ctx, c.Scope, true); err != nil {
		return
	}
	if c.ConversationID == uuid.Nil || c.ThroughMessageID == uuid.Nil {
		return result, domain.ErrInvalid
	}
	err = s.mutate(ctx, func(unit context.Context) error { result, err = s.repo.MarkRead(unit, c); return err })
	return
}

func (s *Service) Search(ctx context.Context, q domain.SearchQuery) (domain.SearchPage, error) {
	if err := s.validateScope(ctx, q.Scope, false); err != nil {
		return domain.SearchPage{}, err
	}
	var err error
	q.Limit, err = boundedLimit(q.Limit, 25, 100)
	if err != nil {
		return domain.SearchPage{}, err
	}
	q.Query = strings.TrimSpace(q.Query)
	if q.ConversationID == uuid.Nil || q.Query == "" || utf8.RuneCountInString(q.Query) > 200 {
		return domain.SearchPage{}, domain.ErrInvalid
	}
	return s.repo.Search(ctx, q)
}

func (s *Service) Resources(ctx context.Context, q domain.SearchQuery) (domain.ResourcePage, error) {
	if err := s.validateScope(ctx, q.Scope, false); err != nil {
		return domain.ResourcePage{}, err
	}
	var err error
	q.Limit, err = boundedLimit(q.Limit, 50, 100)
	if err != nil {
		return domain.ResourcePage{}, err
	}
	switch q.Kind {
	case "":
		q.Kind = "all"
	case "files":
		q.Kind = "file"
	case "links":
		q.Kind = "link"
	case "all", "file", "media", "link":
	default:
		return domain.ResourcePage{}, domain.ErrInvalid
	}
	switch q.Sort {
	case "":
		q.Sort = "newest"
	case "newest", "oldest":
	default:
		return domain.ResourcePage{}, domain.ErrInvalid
	}
	q.Query = strings.TrimSpace(q.Query)
	if q.ConversationID == uuid.Nil || utf8.RuneCountInString(q.Query) > 200 {
		return domain.ResourcePage{}, domain.ErrInvalid
	}
	return s.repo.Resources(ctx, q)
}

func (s *Service) BoundedContext(ctx context.Context, q domain.ContextQuery) ([]domain.Message, error) {
	if err := s.validateScope(ctx, q.Scope, false); err != nil {
		return nil, err
	}
	var err error
	q.Limit, err = boundedLimit(q.Limit, 30, 50)
	if err != nil {
		return nil, err
	}
	if q.ConversationID == uuid.Nil {
		return nil, domain.ErrInvalid
	}
	return s.repo.BoundedContext(ctx, q)
}

func (s *Service) CreateAgentReply(ctx context.Context, c domain.AgentReplyCommand) (result domain.Message, err error) {
	scope := domain.ActorScope{WorkspaceID: c.WorkspaceID, Actor: c.Requester}
	if err = s.validateScope(ctx, scope, true); err != nil {
		return
	}
	if c.Author.Validate() != nil || c.Author.Kind != platformauth.PrincipalSystem || c.Author.PrincipalID != s.mayaID || s.mayaID == uuid.Nil || c.Author.WorkspaceID != c.WorkspaceID || c.Requester.Kind != platformauth.PrincipalHumanUser {
		return result, domain.ErrForbidden
	}
	if c.ThreadRootID == uuid.Nil || c.CreationID == uuid.Nil {
		return result, domain.ErrInvalid
	}
	if err = validateContent(c.Text, c.RichContent, c.MentionedUserIDs, nil, nil); err != nil {
		return
	}
	command := domain.CreateMessageCommand{Scope: scope, ConversationID: c.ConversationID, ThreadRootID: &c.ThreadRootID, CreationID: c.CreationID, Text: c.Text, RichContent: c.RichContent, MentionedUserIDs: c.MentionedUserIDs}
	err = s.mutate(ctx, func(unit context.Context) error {
		conversation, e := s.repo.GetConversation(unit, scope, c.ConversationID, s.mayaID)
		if e != nil {
			return e
		}
		if conversation.TeamID == nil || !c.Author.TeamAccess.Allows(*conversation.TeamID) {
			return domain.ErrForbidden
		}
		result, e = s.repo.CreateMessage(unit, command, c.Author.PrincipalID)
		return e
	})
	return
}

func (s *Service) LinkPendingAttachment(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID, metadata ...domain.AttachmentMetadata) error {
	if err := s.validateAttachmentScope(ctx, scope, id, true); err != nil {
		return err
	}
	var recording domain.AttachmentMetadata
	if len(metadata) > 1 {
		return domain.ErrInvalid
	}
	if len(metadata) == 1 {
		recording = metadata[0]
	}
	if recording.Kind != "" && recording.Kind != "voice-note" {
		return domain.ErrInvalid
	}
	if (recording.Kind == "") != (recording.DurationMS == nil) || (recording.DurationMS != nil && (*recording.DurationMS < 1 || *recording.DurationMS > 300000)) {
		return domain.ErrInvalid
	}
	return s.mutate(ctx, func(unit context.Context) error { return s.repo.LinkPendingAttachment(unit, scope, id, recording) })
}

func (s *Service) AuthorizeAttachment(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID) error {
	if err := s.validateAttachmentScope(ctx, scope, id, false); err != nil {
		return err
	}
	return s.repo.AuthorizeAttachment(ctx, scope, id)
}

func (s *Service) UnlinkPendingAttachment(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID) error {
	if err := s.validateAttachmentScope(ctx, scope, id, true); err != nil {
		return err
	}
	return s.mutate(ctx, func(unit context.Context) error { return s.repo.UnlinkPendingAttachment(unit, scope, id) })
}

func validateAttachmentScope(scope domain.AttachmentScope, id uuid.UUID, write bool) error {
	if err := validateScope(scope.ActorScope, write); err != nil {
		return err
	}
	if scope.ConversationID == uuid.Nil || id == uuid.Nil {
		return domain.ErrInvalid
	}
	return nil
}

func validateContent(text string, rich json.RawMessage, mentions []uuid.UUID, agents []string, attachments []uuid.UUID) error {
	if !utf8.ValidString(text) || utf8.RuneCountInString(text) > domain.MaximumTextRunes || (strings.TrimSpace(text) == "" && len(attachments) == 0) || len(mentions) > 100 || len(attachments) > 10 || len(agents) > 1 {
		return domain.ErrInvalid
	}
	for _, agent := range agents {
		if agent != "maya" {
			return domain.ErrInvalid
		}
	}
	for _, ids := range [][]uuid.UUID{mentions, attachments} {
		seen := make(map[uuid.UUID]bool, len(ids))
		for _, id := range ids {
			if id == uuid.Nil || seen[id] {
				return domain.ErrInvalid
			}
			seen[id] = true
		}
	}
	if len(rich) > 0 && string(rich) != "null" {
		if len(rich) > 64*1024 || !json.Valid(rich) {
			return domain.ErrInvalid
		}
		var value map[string]any
		if json.Unmarshal(rich, &value) != nil || value == nil {
			return fmt.Errorf("%w: rich content must be an object", domain.ErrInvalid)
		}
		if !boundedJSON(value, 0) {
			return fmt.Errorf("%w: rich content is too deeply nested", domain.ErrInvalid)
		}
	}
	return nil
}

func boundedJSON(value any, depth int) bool {
	if depth > 20 {
		return false
	}
	switch node := value.(type) {
	case map[string]any:
		for _, child := range node {
			if !boundedJSON(child, depth+1) {
				return false
			}
		}
	case []any:
		for _, child := range node {
			if !boundedJSON(child, depth+1) {
				return false
			}
		}
	}
	return true
}
