package workflowtemplates

// objectiveRegistry holds objective workflows keyed by template key.
var objectiveRegistry = map[string]struct {
	Statuses   []Status
	DefaultIdx int
}{
	"default": {
		Statuses: []Status{
			{Name: "To Do", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Done", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
			{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
		},
		DefaultIdx: 0,
	},
	"marketing": {
		Statuses: []Status{
			{Name: "To Do", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Done", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
			{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
		},
		DefaultIdx: 0,
	},
	"construction": {
		Statuses: []Status{
			{Name: "To Do", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Done", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
			{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
		},
		DefaultIdx: 0,
	},
}
