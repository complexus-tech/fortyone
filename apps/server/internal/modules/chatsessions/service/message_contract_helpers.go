package chatsessions

import chatsessionsdomain "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"

// cloneJSONValue remains package-local for legacy service tests. Production
// transcript semantics live in the domain package and are shared by the
// service and repository boundaries.
func cloneJSONValue(value any) (any, error) {
	return chatsessionsdomain.CloneJSONValue(value)
}
