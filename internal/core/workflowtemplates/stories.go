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
			{Name: "New Lead", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "Contacted", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Engaged", Category: "started", OrderIndex: 3200, Color: "#eab308"},
			{Name: "Negotiation", Category: "started", OrderIndex: 3400, Color: "#eab308"},
			{Name: "Closed Won", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
			{Name: "Closed Lost", Category: "cancelled", OrderIndex: 6000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"construction": {
		Statuses: []Status{
			{Name: "Planned", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "Scheduled", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Ongoing", Category: "started", OrderIndex: 3500, Color: "#eab308"},
			{Name: "Completed", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
		},
		DefaultIdx: 0,
	},
}
