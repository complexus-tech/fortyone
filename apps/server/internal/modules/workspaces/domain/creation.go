package workspacedomain

import "errors"

type WorkType string

const (
	WorkTypeProduct    WorkType = "product"
	WorkTypeMarketing  WorkType = "marketing"
	WorkTypeOperations WorkType = "operations"
	WorkTypePersonal   WorkType = "personal"
	WorkTypeGeneral    WorkType = "general"
)

var ErrInvalidWorkType = errors.New("invalid workspace work type")

// CreationOptions affect initial content only; they are not workspace settings.
// A nil IncludeExamples preserves the content created for existing clients.
type CreationOptions struct {
	IncludeExamples *bool
	WorkType        WorkType
}

func (options CreationOptions) Validate() error {
	switch options.WorkType {
	case "", WorkTypeProduct, WorkTypeMarketing, WorkTypeOperations, WorkTypePersonal, WorkTypeGeneral:
		return nil
	default:
		return ErrInvalidWorkType
	}
}
