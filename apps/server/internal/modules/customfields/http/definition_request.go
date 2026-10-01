package customfieldshttp

import (
	"encoding/json"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	"github.com/complexus-tech/projects-api/pkg/web"
)

// RawMessage retains omitted versus explicit null without changing legacy
// definition update semantics or bypassing the canonical JSON decoder.
type definitionRequest struct {
	domain.Definition
	Icon json.RawMessage `json:"icon"`
}

func decodeDefinition(r *http.Request) (domain.Definition, error) {
	var request definitionRequest
	if err := web.Decode(r, &request); err != nil {
		return domain.Definition{}, err
	}
	input := request.Definition
	if request.Icon != nil {
		if err := json.Unmarshal(request.Icon, &input.Icon); err != nil {
			return domain.Definition{}, domain.ErrInvalid
		}
		input.IconSet = true
	}
	return input, nil
}
