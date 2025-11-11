package workflowtemplates

// storyRegistry holds story workflows keyed by template key.
var storyRegistry = map[string]struct {
	Statuses   []Status
	DefaultIdx int
}{
	"default": {
		Statuses: []Status{
			{Name: "Planning", Category: "backlog", OrderIndex: 1000, Color: "#6b665c"},
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
			{Name: "Engaged", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "Negotiation", Category: "started", OrderIndex: 5000, Color: "#eab308"},
			{Name: "Closed Won", Category: "completed", OrderIndex: 6000, Color: "#22c55e"},
			{Name: "Closed Lost", Category: "cancelled", OrderIndex: 7000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"construction": {
		Statuses: []Status{
			{Name: "Planned", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "Scheduled", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Ongoing", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "Completed", Category: "completed", OrderIndex: 5000, Color: "#22c55e"},
		},
		DefaultIdx: 0,
	},
	"software": {
		Statuses: []Status{
			{Name: "Backlog", Category: "backlog", OrderIndex: 1000, Color: "#6b665c"},
			{Name: "Todo", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Code Review", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "QA", Category: "started", OrderIndex: 5000, Color: "#eab308"},
			{Name: "Done", Category: "completed", OrderIndex: 6000, Color: "#22c55e"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 7000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"support": {
		Statuses: []Status{
			{Name: "New", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Waiting on Customer", Category: "paused", OrderIndex: 4200, Color: "#64748b"},
			{Name: "Resolved", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
			{Name: "Closed", Category: "completed", OrderIndex: 4100, Color: "#22c55e"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 6000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"recruiting": {
		Statuses: []Status{
			{Name: "Sourced", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "Phone Screen", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Onsite/Panel", Category: "started", OrderIndex: 3200, Color: "#eab308"},
			{Name: "Offer", Category: "started", OrderIndex: 3400, Color: "#eab308"},
			{Name: "Hired", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
			{Name: "Rejected", Category: "cancelled", OrderIndex: 6000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"content": {
		Statuses: []Status{
			{Name: "Idea", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "Drafting", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Editing", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "Scheduled", Category: "started", OrderIndex: 5000, Color: "#eab308"},
			{Name: "Published", Category: "completed", OrderIndex: 6000, Color: "#22c55e"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 7000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"finance": {
		Statuses: []Status{
			{Name: "New Request", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Review", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Approved", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "Processed", Category: "completed", OrderIndex: 5000, Color: "#22c55e"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 6000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"legal": {
		Statuses: []Status{
			{Name: "New Matter", Category: "backlog", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "Intake/Conflicts", Category: "unstarted", OrderIndex: 4200, Color: "#64748b"},
			{Name: "Discovery", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Drafting", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "Review", Category: "started", OrderIndex: 5000, Color: "#eab308"},
			{Name: "Filed/Served", Category: "started", OrderIndex: 6000, Color: "#eab308"},
			{Name: "Hearing/Trial", Category: "started", OrderIndex: 7000, Color: "#eab308"},
			{Name: "Resolved", Category: "completed", OrderIndex: 8000, Color: "#22c55e"},
			{Name: "Closed", Category: "completed", OrderIndex: 9000, Color: "#22c55e"},
			{Name: "Withdrawn", Category: "cancelled", OrderIndex: 10000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"healthcare": {
		Statuses: []Status{
			{Name: "New Case", Category: "backlog", OrderIndex: 1000, Color: "#6b665c"},
			{Name: "Triage", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Assessment", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "Treatment", Category: "started", OrderIndex: 5000, Color: "#eab308"},
			{Name: "Monitoring", Category: "started", OrderIndex: 6000, Color: "#eab308"},
			{Name: "Review", Category: "started", OrderIndex: 7000, Color: "#eab308"},
			{Name: "Discharged", Category: "completed", OrderIndex: 8000, Color: "#22c55e"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 9000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"hr": {
		Statuses: []Status{
			{Name: "New Request", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "In Review", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Manager Approval", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "HR Approval", Category: "started", OrderIndex: 5000, Color: "#eab308"},
			{Name: "Completed", Category: "completed", OrderIndex: 6000, Color: "#22c55e"},
			{Name: "Withdrawn", Category: "cancelled", OrderIndex: 7000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"design": {
		Statuses: []Status{
			{Name: "Intake", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "Research", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Wireframes", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "Visual Design", Category: "started", OrderIndex: 5000, Color: "#eab308"},
			{Name: "Review", Category: "started", OrderIndex: 6000, Color: "#eab308"},
			{Name: "Hand-off", Category: "completed", OrderIndex: 7000, Color: "#22c55e"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 8000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
	"education": {
		Statuses: []Status{
			{Name: "Plan", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
			{Name: "Develop Materials", Category: "started", OrderIndex: 3000, Color: "#eab308"},
			{Name: "Pilot", Category: "started", OrderIndex: 4000, Color: "#eab308"},
			{Name: "Review", Category: "started", OrderIndex: 5000, Color: "#eab308"},
			{Name: "Deliver", Category: "started", OrderIndex: 6000, Color: "#eab308"},
			{Name: "Completed", Category: "completed", OrderIndex: 7000, Color: "#22c55e"},
			{Name: "Cancelled", Category: "cancelled", OrderIndex: 8000, Color: "#f43f5e"},
		},
		DefaultIdx: 0,
	},
}
