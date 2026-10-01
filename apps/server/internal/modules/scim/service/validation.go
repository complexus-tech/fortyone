package scimservice

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
)

func invalid(detail string) error { return fmt.Errorf("%w: %s", domain.ErrInvalidInput, detail) }
func validate(in domain.Input, creating bool) (domain.Mutation, error) {
	if len(in.Schemas) == 0 || len(in.Schemas) > 2 {
		return domain.Mutation{}, invalid("schemas must include the User schema")
	}
	seen := map[string]bool{}
	for _, schema := range in.Schemas {
		if seen[schema] || (schema != domain.UserSchema && schema != domain.EnterpriseSchema) {
			return domain.Mutation{}, invalid("unsupported schema")
		}
		seen[schema] = true
	}
	if !seen[domain.UserSchema] || (in.Enterprise != nil && !seen[domain.EnterpriseSchema]) {
		return domain.Mutation{}, invalid("missing resource schema")
	}
	in.UserName = strings.TrimSpace(in.UserName)
	if utf8.RuneCountInString(in.UserName) < 1 || utf8.RuneCountInString(in.UserName) > 255 || strings.ContainsAny(in.UserName, "\r\n\x00") {
		return domain.Mutation{}, invalid("userName is required and must be at most 255 characters")
	}
	if in.ExternalID != nil {
		value := strings.TrimSpace(*in.ExternalID)
		if value == "" {
			in.ExternalID = nil
		} else if len(value) > 255 {
			return domain.Mutation{}, invalid("externalId is too long")
		} else {
			in.ExternalID = &value
		}
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	profile := domain.Profile{DisplayName: in.DisplayName, Name: in.Name, Emails: in.Emails, Enterprise: in.Enterprise}
	body, err := json.Marshal(profile)
	if err != nil {
		return domain.Mutation{}, err
	}
	if len(body) > 16384 {
		return domain.Mutation{}, invalid("user profile is too large")
	}
	if len(in.DisplayName) > 255 || len(in.Emails) > 10 {
		return domain.Mutation{}, invalid("user profile exceeds the supported bounds")
	}
	email := ""
	primary := false
	for i, item := range in.Emails {
		parsed, err := mail.ParseAddress(item.Value)
		if err != nil || parsed.Address != item.Value || len(item.Value) > 255 {
			return domain.Mutation{}, invalid("emails must contain valid addresses")
		}
		if len(item.Type) > 32 {
			return domain.Mutation{}, invalid("email type is too long")
		}
		if item.Primary {
			if primary {
				return domain.Mutation{}, invalid("only one primary email is allowed")
			}
			primary = true
			email = strings.ToLower(parsed.Address)
		} else if i == 0 {
			email = strings.ToLower(parsed.Address)
		}
	}
	if email == "" {
		parsed, err := mail.ParseAddress(in.UserName)
		if err == nil && parsed.Address == in.UserName {
			email = strings.ToLower(parsed.Address)
		}
	}
	if creating && email == "" {
		return domain.Mutation{}, invalid("a primary email address is required to create a member")
	}
	fullName := in.DisplayName
	if fullName == "" && in.Name != nil {
		fullName = in.Name.Formatted
		if fullName == "" {
			fullName = strings.TrimSpace(in.Name.GivenName + " " + in.Name.FamilyName)
		}
	}
	if len(fullName) > 255 {
		fullName = in.UserName
	}
	return domain.Mutation{UserName: in.UserName, ExternalID: in.ExternalID, Active: active, Profile: body, Email: email, FullName: fullName}, nil
}
