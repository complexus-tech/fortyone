package customfieldsdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("custom field resource was not found")
	ErrForbidden = errors.New("custom field permission is required")
	ErrInvalid   = errors.New("custom field input is invalid")
	ErrConflict  = errors.New("custom field resource changed; reload and try again")
	ErrLimit     = errors.New("custom field limit reached")
)

const (
	MaxActiveFields = 50
	MaxFields       = 200
	MaxOptions      = 100
)

type FieldType string

const (
	Text   FieldType = "text"
	Number FieldType = "number"
	Money  FieldType = "money"
	Date   FieldType = "date"
	Select FieldType = "select"
	Person FieldType = "person"
)

type Scope struct{ ActorID, WorkspaceID, TeamID uuid.UUID }
type Option struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	ArchivedAt *time.Time `json:"archivedAt"`
}
type Field struct {
	ID           uuid.UUID  `json:"id"`
	TeamID       uuid.UUID  `json:"teamId"`
	Name         string     `json:"name"`
	Type         FieldType  `json:"type"`
	Currency     *string    `json:"currency"`
	Icon         *string    `json:"icon"`
	Options      []Option   `json:"options"`
	ShowOnCreate bool       `json:"showOnCreate"`
	ArchivedAt   *time.Time `json:"archivedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
type Definition struct {
	Name              string        `json:"name" validate:"required,max=100"`
	Type              FieldType     `json:"type" validate:"omitempty,oneof=text number money date select person"`
	Currency          *string       `json:"currency" validate:"omitempty,len=3"`
	Icon              *string       `json:"icon"`
	IconSet           bool          `json:"-"`
	Options           []OptionInput `json:"options" validate:"max=100,dive"`
	ShowOnCreate      bool          `json:"showOnCreate"`
	ExpectedUpdatedAt *time.Time    `json:"expectedUpdatedAt"`
}
type OptionInput struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name" validate:"required,max=100"`
}
type Value struct {
	FieldID uuid.UUID `json:"fieldId"`
	Value   *string   `json:"value" validate:"omitempty,max=4000"`
}
type Snapshot struct {
	Fields  []Field `json:"fields"`
	Values  []Value `json:"values"`
	Version int64   `json:"version"`
}

type BatchInput struct {
	StoryIDs []uuid.UUID `json:"storyIds" validate:"required,min=1,max=100"`
}

type BatchStory struct {
	StoryID uuid.UUID `json:"storyId"`
	Values  []Value   `json:"values"`
	Version int64     `json:"version"`
}

type BatchSnapshot struct {
	Items []BatchStory `json:"items"`
}
type ValuePatch struct {
	Values          []Value `json:"values" validate:"max=50,dive"`
	ExpectedVersion *int64  `json:"expectedVersion" validate:"omitempty,gte=0"`
}
type ReportInput struct {
	FieldID     uuid.UUID   `json:"fieldId"`
	Aggregation string      `json:"aggregation" validate:"required,oneof=sum average min max count"`
	GroupBy     string      `json:"groupBy" validate:"required,oneof=none status assignee month"`
	StatusIDs   []uuid.UUID `json:"statusIds" validate:"max=100"`
	AssigneeIDs []uuid.UUID `json:"assigneeIds" validate:"max=100"`
	StartDate   *string     `json:"startDate" validate:"omitempty,datetime=2006-01-02"`
	EndDate     *string     `json:"endDate" validate:"omitempty,datetime=2006-01-02"`
	DateBasis   string      `json:"dateBasis" validate:"required,max=36"`
}
type ReportRow struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	// Value is an exact decimal string. An empty average/min/max group uses
	// "" with Count == 0; sum/count use "0" instead.
	Value string `json:"value"`
	Count int64  `json:"count"`
}
type Report struct {
	Field        Field       `json:"field"`
	Aggregation  string      `json:"aggregation"`
	GroupBy      string      `json:"groupBy"`
	Currency     *string     `json:"currency"`
	TotalCount   int64       `json:"totalCount"`
	ValuedCount  int64       `json:"valuedCount"`
	MissingCount int64       `json:"missingCount"`
	Rows         []ReportRow `json:"rows"`
}
