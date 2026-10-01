package customfieldsdomain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTypedValuesPreserveDecimalsAndDistinguishMissing(t *testing.T) {
	for _, input := range []string{"0", "0.00", "-12.250000000000000001", "9007199254740993.01"} {
		value, err := NormalizeValue(Field{Type: Money}, &input)
		if err != nil || value == nil || *value != input {
			t.Fatalf("value %q = %v, %v", input, value, err)
		}
	}
	for _, input := range []string{"", "1e3", "NaN", "Infinity", "01", "+1", "0.", strings.Repeat("9", 39), "0." + strings.Repeat("1", 19)} {
		if _, err := NormalizeValue(Field{Type: Number}, &input); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid %q accepted: %v", input, err)
		}
	}
	if value, err := NormalizeValue(Field{Type: Money}, nil); err != nil || value != nil {
		t.Fatalf("null = %v, %v", value, err)
	}
	empty := ""
	if value, err := NormalizeValue(Field{Type: Text}, &empty); err != nil || value == nil || *value != "" {
		t.Fatal("empty text was conflated with null")
	}
}

func TestDefinitionCurrencyAndOptionsAreStable(t *testing.T) {
	usd, eur := "USD", "EUR"
	existing := Field{Type: Money, Currency: &usd}
	if err := ValidateDefinition(Definition{Name: "Value", Type: Money, Currency: &eur}, &existing); !errors.Is(err, ErrInvalid) {
		t.Fatalf("currency mutation = %v", err)
	}
	if err := ValidateDefinition(Definition{Name: "Value", Type: Text}, &existing); !errors.Is(err, ErrInvalid) {
		t.Fatalf("type mutation = %v", err)
	}
	if err := ValidateDefinition(Definition{Name: "Value", Type: Money, Currency: &usd}, nil); err != nil {
		t.Fatal(err)
	}
	for _, currency := range []string{"usd", "ZZZ", "US", "XXX!"} {
		if err := ValidateDefinition(Definition{Name: "Value", Type: Money, Currency: &currency}, nil); !errors.Is(err, ErrInvalid) {
			t.Fatalf("bad currency %q = %v", currency, err)
		}
	}
	if err := ValidateDefinition(Definition{Name: "Stage", Type: Select, Options: []OptionInput{{Name: "New"}, {Name: "new"}}}, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("duplicate choices = %v", err)
	}
	optionID := uuid.New()
	choice := optionID.String()
	now := time.Now()
	field := Field{Type: Select, Options: []Option{{ID: optionID, Name: "Old", ArchivedAt: &now}}}
	if _, err := NormalizeValue(field, &choice); !errors.Is(err, ErrInvalid) {
		t.Fatalf("archived option = %v", err)
	}
}

func TestDatesAndReportRangeValidation(t *testing.T) {
	for _, input := range []string{"2026-02-29", "2026-1-01", "2026-01-01T00:00:00Z"} {
		if _, err := NormalizeValue(Field{Type: Date}, &input); !errors.Is(err, ErrInvalid) {
			t.Fatalf("date %q = %v", input, err)
		}
	}
	input := ReportInput{FieldID: uuid.New(), Aggregation: "sum", GroupBy: "month", DateBasis: uuid.NewString()}
	if err := ValidateReport(input); err != nil {
		t.Fatal(err)
	}
	start, end := "2026-04-02", "2026-04-01"
	input.StartDate, input.EndDate = &start, &end
	if err := ValidateReport(input); !errors.Is(err, ErrInvalid) {
		t.Fatalf("reversed range = %v", err)
	}
}

func TestDefinitionIconsUseTheCanonicalCatalog(t *testing.T) {
	for _, icon := range []string{
		"text", "number", "calendar", "list", "person", "team", "workspace", "goal", "star", "checklist", "link", "email", "clock", "work", "attachment", "globe",
		"tag", "pin", "objective", "strategy", "roadmap", "home", "health", "approval", "chat", "comment", "share", "image", "video", "microphone", "book", "help", "info", "analytics", "dashboard", "workflow", "kanban", "sprint", "automation", "history", "lock", "key", "code", "warning",
		"money", "coins", "wallet", "credit-card", "bank", "invoice", "percent", "calculator", "piggy-bank", "target-money", "building", "megaphone", "store", "package", "phone", "handshake", "shopping-cart",
	} {
		if err := ValidateDefinition(Definition{Name: "Label", Type: Text, Icon: &icon}, nil); err != nil {
			t.Fatal(icon, err)
		}
	}
	for _, icon := range []string{"", "Star", "star ", "<svg>", "unknown"} {
		if err := ValidateDefinition(Definition{Name: "Label", Type: Text, Icon: &icon}, nil); !errors.Is(err, ErrInvalid) {
			t.Fatal("unsupported icon accepted", icon, err)
		}
	}
}
