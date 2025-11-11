package workflowtemplates

// Status represents a workflow status used during seeding.
type Status struct {
	Name       string
	Category   string // backlog | unstarted | started | paused | completed | cancelled
	OrderIndex int
	Color      string
}

const (
	EntityStory     = "story"
	EntityObjective = "objective"
)

// GetWorkflow returns the statuses and default index for a template key and entity.
// If unknown, ok=false; callers can fallback to ("default", entity).
func GetWorkflow(key, entity string) ([]Status, int, bool) {
	switch entity {
	case EntityStory:
		if set, ok := storyRegistry[key]; ok {
			return set.Statuses, set.DefaultIdx, true
		}
	case EntityObjective:
		if set, ok := objectiveRegistry[key]; ok {
			return set.Statuses, set.DefaultIdx, true
		}
	}
	return nil, -1, false
}
