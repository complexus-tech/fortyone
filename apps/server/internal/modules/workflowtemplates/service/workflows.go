package workflowtemplates

// Status represents a workflow status used during seeding.
type Status struct {
	Name       string
	Category   string // backlog | unstarted | started | paused | completed | cancelled
	OrderIndex int
	Color      string
}

// GetWorkflow returns the statuses and default index for a template key.
// If unknown, ok=false; callers can fallback to ("default", entity).
func GetWorkflow(key string) ([]Status, int, bool) {
	if set, ok := storyRegistry[key]; ok {
		return set.Statuses, set.DefaultIdx, true
	}
	return nil, -1, false
}
