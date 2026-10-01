package customfieldsdomain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/currency"
)

var decimalPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,37})(\.[0-9]{1,18})?$`)

func ValidateDefinition(input Definition, existing *Field) error {
	if strings.TrimSpace(input.Name) == "" || len(input.Name) > 100 {
		return fmt.Errorf("%w: name is required", ErrInvalid)
	}
	if input.Icon != nil {
		switch *input.Icon {
		case "text", "number", "calendar", "list", "person", "team", "workspace", "goal", "star", "checklist", "link", "email", "clock", "work", "attachment", "globe",
			"tag", "pin", "objective", "strategy", "roadmap", "home", "health", "approval", "chat", "comment", "share", "image", "video", "microphone", "book", "help", "info", "analytics", "dashboard", "workflow", "kanban", "sprint", "automation", "history", "lock", "key", "code", "warning",
			"money", "coins", "wallet", "credit-card", "bank", "invoice", "percent", "calculator", "piggy-bank", "target-money", "building", "megaphone", "store", "package", "phone", "handshake", "shopping-cart":
		default:
			return fmt.Errorf("%w: unsupported field icon", ErrInvalid)
		}
	}
	fieldType := input.Type
	if existing != nil {
		if fieldType == "" {
			fieldType = existing.Type
		}
		if fieldType != existing.Type {
			return fmt.Errorf("%w: field type is immutable", ErrInvalid)
		}
		if input.Currency != nil && (existing.Currency == nil || *input.Currency != *existing.Currency) {
			return fmt.Errorf("%w: currency is immutable", ErrInvalid)
		}
		if existing.ArchivedAt != nil {
			return ErrConflict
		}
	}
	switch fieldType {
	case Text, Number, Money, Date, Select, Person:
	default:
		return fmt.Errorf("%w: unsupported field type", ErrInvalid)
	}
	if existing == nil {
		if fieldType == Money {
			if input.Currency == nil || len(*input.Currency) != 3 || strings.ToUpper(*input.Currency) != *input.Currency {
				return fmt.Errorf("%w: ISO currency is required", ErrInvalid)
			}
			if _, err := currency.ParseISO(*input.Currency); err != nil {
				return fmt.Errorf("%w: unsupported currency", ErrInvalid)
			}
		} else if input.Currency != nil {
			return fmt.Errorf("%w: currency is only available for money", ErrInvalid)
		}
	}
	if len(input.Options) > MaxOptions || (fieldType != Select && len(input.Options) != 0) {
		return fmt.Errorf("%w: options do not match field type", ErrInvalid)
	}
	names, ids := map[string]bool{}, map[uuid.UUID]bool{}
	for _, option := range input.Options {
		name := strings.ToLower(strings.TrimSpace(option.Name))
		if name == "" || len(option.Name) > 100 || names[name] || (option.ID != uuid.Nil && ids[option.ID]) {
			return fmt.Errorf("%w: select options must be unique", ErrInvalid)
		}
		names[name] = true
		ids[option.ID] = true
	}
	return nil
}

// NormalizeValue preserves exact decimal strings. Empty text is a value;
// clearing any typed value requires an explicit null.
func NormalizeValue(field Field, value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized := *value
	switch field.Type {
	case Text:
		if len(normalized) > 4000 {
			return nil, fmt.Errorf("%w: text exceeds 4000 bytes", ErrInvalid)
		}
	case Number, Money:
		if !decimalPattern.MatchString(normalized) {
			return nil, fmt.Errorf("%w: use a plain decimal with up to 38 integral and 18 fractional digits", ErrInvalid)
		}
	case Date:
		parsed, err := time.Parse(time.DateOnly, normalized)
		if err != nil || parsed.Format(time.DateOnly) != normalized {
			return nil, fmt.Errorf("%w: date must use YYYY-MM-DD", ErrInvalid)
		}
	case Select, Person:
		id, err := uuid.Parse(normalized)
		if err != nil || id == uuid.Nil {
			return nil, fmt.Errorf("%w: value must be a resource ID", ErrInvalid)
		}
		normalized = id.String()
		if field.Type == Select {
			valid := false
			for _, option := range field.Options {
				if option.ID == id && option.ArchivedAt == nil {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("%w: select option is unavailable", ErrInvalid)
			}
		}
	default:
		return nil, ErrInvalid
	}
	return &normalized, nil
}

func ValidatePatch(patch ValuePatch) error {
	if len(patch.Values) > MaxActiveFields {
		return ErrLimit
	}
	if patch.ExpectedVersion != nil && *patch.ExpectedVersion < 0 {
		return ErrInvalid
	}
	seen := map[uuid.UUID]bool{}
	for _, value := range patch.Values {
		if value.FieldID == uuid.Nil || seen[value.FieldID] {
			return fmt.Errorf("%w: field IDs must be unique", ErrInvalid)
		}
		seen[value.FieldID] = true
	}
	return nil
}

func ValidateReport(input ReportInput) error {
	if input.FieldID == uuid.Nil {
		return ErrInvalid
	}
	switch input.Aggregation {
	case "sum", "average", "min", "max", "count":
	default:
		return ErrInvalid
	}
	switch input.GroupBy {
	case "none", "status", "assignee", "month":
	default:
		return ErrInvalid
	}
	if input.DateBasis != "created" && input.DateBasis != "completed" {
		if id, err := uuid.Parse(input.DateBasis); err != nil || id == uuid.Nil {
			return ErrInvalid
		}
	}
	if len(input.StatusIDs) > 100 || len(input.AssigneeIDs) > 100 {
		return ErrInvalid
	}
	for _, date := range []*string{input.StartDate, input.EndDate} {
		if date != nil {
			if _, err := time.Parse(time.DateOnly, *date); err != nil {
				return ErrInvalid
			}
		}
	}
	if input.StartDate != nil && input.EndDate != nil && *input.StartDate > *input.EndDate {
		return fmt.Errorf("%w: date range is reversed", ErrInvalid)
	}
	return nil
}
