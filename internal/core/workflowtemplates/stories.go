package workflowtemplates

// storyRegistry holds story workflows keyed by template key.
var storyRegistry = map[string]struct {
	Statuses   []Status
	DefaultIdx int
}{
	"default": {
		Statuses: []Status{
			{Name: "Backlog", Category: "backlog", OrderIndex: 1000, Color: "#6b665c"},
			{Name: "To Do", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Done", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
			{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 6000, Color: "#f43f5e"},
		},
		DefaultIdx: 1,
	},
	"marketing": {
		Statuses: []Status{
			{Name: "Backlog", Category: "backlog", OrderIndex: 1000, Color: "#6b665c"},
			{Name: "Briefing", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Design", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Review", Category: "paused", OrderIndex: 3500, Color: "#64748b"},
			{Name: "Scheduled", Category: "started", OrderIndex: 3600, Color: "#a3e635"},
			{Name: "Published", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
			{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 6000, Color: "#f43f5e"},
		},
		DefaultIdx: 1,
	},
	"construction": {
		Statuses: []Status{
			{Name: "Backlog", Category: "backlog", OrderIndex: 1000, Color: "#6b665c"},
			{Name: "Procurement", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Inspection", Category: "paused", OrderIndex: 3500, Color: "#64748b"},
			{Name: "Completed", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
			{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 6000, Color: "#f43f5e"},
		},
		DefaultIdx: 1,
	},
}
