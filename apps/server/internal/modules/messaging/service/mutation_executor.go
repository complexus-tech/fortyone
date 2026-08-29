package messaging

import (
	"crypto/rand"
	"crypto/sha256"
	"io"
	"time"

	"github.com/google/uuid"
)

const (
	storyMutationConfirmationKind = "story_mutation_confirmation_required"
	storyMutationTokenVersion     = 1
	storyMutationConfirmationTTL  = 10 * time.Minute
	maximumStoryMutationTokenSize = 2_000
	maximumStoryTitleRunes        = 255
	maximumBatchStoryCount        = 10
	maximumBatchDescriptionRunes  = 2_000
	batchStoryProposalVersion     = 1
	batchStoryTokenBytes          = 32

	storyMutationStatusApplied        = "applied"
	storyMutationStatusAlreadyApplied = "already_applied"
	storyMutationStatusPartial        = "partial"

	assigneeActionUnchanged  = "unchanged"
	assigneeActionMe         = "me"
	assigneeActionNamed      = "named"
	assigneeActionUnassigned = "unassigned"

	storyTimeActionUnchanged = "unchanged"
	storyTimeActionSet       = "set"
	storyTimeActionClear     = "clear"
)

var storyPriorities = map[string]struct{}{
	"No Priority": {},
	"Low":         {},
	"Medium":      {},
	"High":        {},
	"Urgent":      {},
}

type storyMutationExecutor struct {
	stories StoryMutationService
	store   StoryMutationConfirmationStore
	key     []byte
	now     func() time.Time
	random  io.Reader
}

type storyMutationClaims struct {
	Version                  int                    `json:"v"`
	ConfirmationID           uuid.UUID              `json:"i"`
	Operation                StoryMutationOperation `json:"o"`
	WorkspaceID              uuid.UUID              `json:"w"`
	UserID                   uuid.UUID              `json:"u"`
	TeamID                   uuid.UUID              `json:"t"`
	StoryID                  *uuid.UUID             `json:"s,omitempty"`
	ExpectedUpdatedAt        *time.Time             `json:"e,omitempty"`
	Title                    *string                `json:"n,omitempty"`
	Priority                 *string                `json:"p,omitempty"`
	AssigneeAction           string                 `json:"a"`
	StatusID                 *uuid.UUID             `json:"st,omitempty"`
	SprintID                 *uuid.UUID             `json:"sp,omitempty"`
	ObjectiveID              *uuid.UUID             `json:"oj,omitempty"`
	KeyResultID              *uuid.UUID             `json:"k,omitempty"`
	StartDate                *time.Time             `json:"sd,omitempty"`
	EndDate                  *time.Time             `json:"ed,omitempty"`
	EstimatedDurationAction  string                 `json:"da,omitempty"`
	EstimatedDurationMinutes *int                   `json:"du,omitempty"`
	MinimumFocusBlockAction  string                 `json:"fa,omitempty"`
	MinimumFocusBlockMinutes *int                   `json:"fb,omitempty"`
	AutoSchedulingEnabled    *bool                  `json:"ae,omitempty"`
	AutoSchedulingLocked     *bool                  `json:"al,omitempty"`
	Comment                  *string                `json:"c,omitempty"`
	MentionIDs               []uuid.UUID            `json:"m,omitempty"`
	LabelIDs                 []uuid.UUID            `json:"l,omitempty"`
	RelationStoryID          *uuid.UUID             `json:"r,omitempty"`
	RelationType             string                 `json:"rt,omitempty"`
	ExpiresAt                time.Time              `json:"x"`
}

type storyTimeMutation struct {
	estimatedDurationAction  string
	estimatedDurationMinutes *int
	minimumFocusBlockAction  string
	minimumFocusBlockMinutes *int
}

type storyAutoSchedulingMutation struct {
	enabled *bool
	locked  *bool
}

type storyMutationConfirmationToolResult struct {
	Kind         string                    `json:"kind"`
	Confirmation StoryMutationConfirmation `json:"confirmation"`
}

type batchStoryMutationProposal struct {
	Version   int                      `json:"version"`
	SourceURL string                   `json:"source_url,omitempty"`
	Items     []batchStoryMutationItem `json:"items"`
}

type batchStoryMutationItem struct {
	Title                    string     `json:"title"`
	Description              string     `json:"description,omitempty"`
	Priority                 string     `json:"priority"`
	AssigneeID               *uuid.UUID `json:"assignee_id,omitempty"`
	EstimatedDurationMinutes *int       `json:"estimated_duration_minutes,omitempty"`
	MinimumFocusBlockMinutes *int       `json:"minimum_focus_block_minutes,omitempty"`
	AutoSchedulingEnabled    bool       `json:"auto_scheduling_enabled"`
}

func newStoryMutationExecutor(
	storiesService StoryMutationService,
	secret string,
	store StoryMutationConfirmationStore,
) *storyMutationExecutor {
	key := sha256.Sum256([]byte("fortyone:messaging:story-mutation:v1\x00" + secret))
	return &storyMutationExecutor{
		stories: storiesService,
		store:   store,
		key:     append([]byte(nil), key[:]...),
		now:     time.Now,
		random:  rand.Reader,
	}
}
