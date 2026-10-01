package scimservice

import (
	"encoding/json"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
)

func TestUserValidationRequiresSafeSchemaAndCreationIdentity(t *testing.T) {
	for _, input := range []domain.Input{
		{Schemas: []string{domain.UserSchema}, UserName: "not-an-email"},
		{Schemas: []string{domain.UserSchema}, UserName: "member@example.com", Emails: []domain.Email{{Value: "one@example.com", Primary: true}, {Value: "two@example.com", Primary: true}}},
		{Schemas: []string{"unknown"}, UserName: "member@example.com"},
		{Schemas: []string{domain.UserSchema}, UserName: "member@example.com", Enterprise: &domain.Enterprise{Department: "Finance"}},
	} {
		if _, err := validate(input, true); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatal("invalid create accepted", input, err)
		}
	}
	mutation, err := validate(domain.Input{Schemas: []string{domain.UserSchema}, UserName: "DirectoryId", Emails: []domain.Email{{Value: "Actual@example.com", Primary: true}}}, true)
	if err != nil || mutation.Email != "actual@example.com" || !mutation.Active {
		t.Fatal(mutation, err)
	}
}
func TestPatchComplexMergeAndRequiredMutability(t *testing.T) {
	values := map[string]json.RawMessage{"name": json.RawMessage(`{"givenName":"Original","familyName":"Kept"}`), "userName": json.RawMessage(`"member@example.com"`)}
	if err := applyPatch(values, domain.Operation{Op: "replace", Path: "NAME.GIVENNAME", Value: json.RawMessage(`"Changed"`)}); err != nil {
		t.Fatal(err)
	}
	var name domain.Name
	if err := json.Unmarshal(values["name"], &name); err != nil || name.GivenName != "Changed" || name.FamilyName != "Kept" {
		t.Fatal(name, err)
	}
	if err := applyPatch(values, domain.Operation{Op: "remove", Path: "userName"}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatal("required attribute removed", err)
	}
	if err := applyPatch(values, domain.Operation{Op: "replace", Path: "roles", Value: json.RawMessage(`["admin"]`)}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatal("provider controlled tenant role", err)
	}
	if err := applyPatch(values, domain.Operation{Op: "replace", Value: json.RawMessage(`{"active":false,"displayName":"Profile"}`)}); err != nil {
		t.Fatal(err)
	}
	if string(values["active"]) != "false" {
		t.Fatal("pathless update missing")
	}
}

func TestPatchEmailValuePathKeepsOtherValues(t *testing.T) {
	values := map[string]json.RawMessage{"emails": json.RawMessage(`[{"value":"home@example.com","type":"home"},{"value":"old@example.com","type":"work","primary":true}]`)}
	matched, err := patchEmails(values, domain.Operation{Op: "replace", Path: `emails[type eq "work"].value`, Value: json.RawMessage(`"new@example.com"`)}, "replace")
	var emails []domain.Email
	_ = json.Unmarshal(values["emails"], &emails)
	if !matched || err != nil || len(emails) != 2 || emails[0].Value != "home@example.com" || emails[1].Value != "new@example.com" || !emails[1].Primary {
		t.Fatal(emails, matched, err)
	}
	if _, err = patchEmails(values, domain.Operation{Path: `emails[type eq "missing"].value`, Value: json.RawMessage(`"new@example.com"`)}, "replace"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatal("missing target accepted", err)
	}
}
