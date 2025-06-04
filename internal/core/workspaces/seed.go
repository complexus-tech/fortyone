package workspaces

import (
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/google/uuid"
)

func seedStories(teamID uuid.UUID, userID uuid.UUID, statuses []states.CoreState) []stories.CoreNewStory {
	storyData := []struct {
		title           string
		description     string
		descriptionHTML string
		priority        string
		statusCategory  string
	}{
		{
			"Welcome to Complexus: Streamline Your Project Management",
			"Welcome to Complexus! Our platform is designed to help your team build better products faster by streamlining workflows, tracking objectives and stories precisely, and facilitating seamless collaboration.",
			`<p>Welcome to Complexus! Our platform is designed to help your team build better products faster by streamlining workflows, tracking objectives and stories precisely, and facilitating seamless collaboration.</p>
			
			<h3>Key Features</h3>
			<ul>
				<li><strong>Plan with Clarity:</strong> Define goals, map out objectives, and organize work into manageable stories.</li>
				<li><strong>Track Progress:</strong> Capture objectives and stories using flexible tracking. Monitor real-time progress with customizable boards and lists.</li>
				<li><strong>Collaborate Effectively:</strong> Keep your team aligned with shared views, discussions, and notifications.</li>
				<li><strong>Continuous Improvement:</strong> Leverage insights from your work to refine processes and boost team velocity.</li>
			</ul>
			
			<p>To get started, explore our <a href="https://docs.complexus.app/concepts" target="_blank">Core Concepts</a> to understand the fundamental building blocks of Complexus.</p>`,
			"High",
			"completed",
		},
		{
			"Setting Up Your Workspace in Complexus",
			"Creating a well-structured workspace is the first step to effective project management in Complexus.",
			`<p>Creating a well-structured workspace is the first step to effective project management in Complexus.</p>
			
			<h3>Steps to Set Up Your Workspace</h3>
			<ol>
				<li><strong>Create a New Workspace:</strong> Go to the menu to the top left corner and click on 'Create workspace'. Provide a name for your workspace.</li>
				<li><strong>Configure Settings:</strong> Customize your workspace by setting up roles, permissions, and preferences to suit your team's needs.</li>
				<li><strong>Invite Team Members:</strong> Add colleagues to your workspace to start collaborating on objectives.</li>
			</ol>
			
			<p>For detailed instructions, refer to our guide on <a href="https://docs.complexus.app/start-here/set-up-your-workspace" target="_blank">Setting Up Your Workspace</a>.</p>`,
			"High",
			"started",
		},
		{
			"Understanding Roles and Permissions in Complexus",
			"Effective collaboration requires clear roles and permissions within your workspace.",
			`<p>Effective collaboration requires clear roles and permissions within your workspace.</p>
			
			<h3>Key Points</h3>
			<ul>
				<li><strong>Roles:</strong> Assign predefined roles to team members based on their responsibilities.</li>
				<li><strong>Permissions:</strong> Control access to objectives, stories, and settings to ensure data security and appropriate information sharing.</li>
			</ul>
			
			<p>Learn more about configuring roles and permissions in our <a href="https://docs.complexus.app/roles-and-permissions" target="_blank">Roles and Permissions</a> section.</p>`,
			"Medium",
			"backlog",
		},
		{
			"Navigating Core Features of Complexus",
			"Complexus offers a suite of tools to enhance your project management experience.",
			`<p>Complexus offers a suite of tools to enhance your project management experience.</p>
			
			<h3>Core Features</h3>
			<ul>
				<li><strong>Objectives:</strong> Capture and prioritize objectives to maintain a clear backlog.</li>
				<li><strong>Stories:</strong> Capture and prioritize stories to maintain a clear backlog.</li>
				<li><strong>Customizable Boards:</strong> Visualize your workflow with boards tailored to your team's processes.</li>
			</ul>
			
			<p>Explore these features in detail in our <a href="https://docs.complexus.app/product-guide/stories" target="_blank">Core Features</a> section.</p>`,
			"Medium",
			"started",
		},
		{
			"Accessing Help and Support in Complexus",
			"Should you need assistance while using Complexus, our support resources are readily available.",
			`<p>Should you need assistance while using Complexus, our support resources are readily available.</p>
			
			<h3>Support Resources</h3>
			<ul>
				<li><strong>Frequently Asked Questions:</strong> Find answers to common queries in our <a href="https://docs.complexus.app/help-and-support/faq" target="_blank">FAQ section</a>.</li>
				<li><strong>Contact Support:</strong> Reach out to our support team for personalized assistance through the <a href="https://docs.complexus.app/help-and-support/contact-us" target="_blank">Get Support page</a>.</li>
			</ul>
			
			<p>We're committed to ensuring you have a smooth experience with Complexus.</p>`,
			"Low",
			"unstarted",
		},
	}

	// Find statuses by category
	statusMap := make(map[string]uuid.UUID)
	for _, status := range statuses {
		statusMap[status.Category] = status.ID
	}

	// Fallback to first status if specific category not found
	fallbackStatusID := statuses[0].ID

	s := make([]stories.CoreNewStory, len(storyData))
	for i, data := range storyData {
		desc := data.description
		descHTML := data.descriptionHTML

		// Get status ID by category, fallback if not found
		statusID := statusMap[data.statusCategory]
		if statusID == uuid.Nil {
			statusID = fallbackStatusID
		}

		s[i] = stories.CoreNewStory{
			Title:           data.title,
			Description:     &desc,
			DescriptionHTML: &descHTML,
			Reporter:        &userID,
			Assignee:        &userID,
			Priority:        data.priority,
			Team:            teamID,
			Status:          &statusID,
		}
	}

	return s
}
