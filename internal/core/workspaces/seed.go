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
	}{
		{
			"Welcome to Your New Workspace",
			"Get started with your new workspace and learn about the key features that will help your team succeed.",
			`<h1>Welcome to Your New Workspace</h1>
      <p>We're excited to help you and your team achieve your goals. This workspace will be your central hub for managing objectives, tracking progress, and collaborating with your team.</p>
      
      <h2>What You Can Do Here</h2>
      <ul>
        <li>Set and track Objectives & Key Results (OKRs)</li>
        <li>Manage team sprints and track progress</li>
        <li>Create and assign stories to team members</li>
        <li>Collaborate with your team in real-time</li>
      </ul>

      <h2>Getting Started</h2>
      <p>Here are the first steps to set up your workspace for success:</p>
      <ol>
        <li>
          <strong>Invite Your Team</strong>
          <ul>
            <li>Click on the "Settings" in the sidebar</li>
            <li>Navigate to "Team Members"</li>
            <li>Use the "Invite Members" button to add your colleagues</li>
            <li>Assign appropriate roles (Admin, Member, or Viewer)</li>
          </ul>
        </li>
        <li>
          <strong>Create Teams</strong>
          <ul>
            <li>Set up teams for different departments or projects</li>
            <li>Assign team leads and members</li>
            <li>Define team responsibilities and focus areas</li>
          </ul>
        </li>
        <li>
          <strong>Set Up Your First OKR</strong>
          <ul>
            <li>Navigate to the OKRs section</li>
            <li>Create an objective for this quarter</li>
            <li>Add measurable key results</li>
          </ul>
        </li>
        <li>
          <strong>Plan Your First Sprint</strong>
          <ul>
            <li>Go to the Sprints section</li>
            <li>Create a new sprint with a clear goal</li>
            <li>Add relevant stories and assign them</li>
          </ul>
        </li>
      </ol>

      <h2>Workspace Organization</h2>
      <p>Your workspace is organized into several key areas:</p>
      <ul>
        <li>
          <strong>Dashboard:</strong> Get an overview of all activities, recent updates, and important metrics
        </li>
        <li>
          <strong>OKRs:</strong> Manage your objectives and key results, track progress, and align team goals
        </li>
        <li>
          <strong>Sprints:</strong> Plan and execute work in focused iterations
        </li>
        <li>
          <strong>Stories:</strong> Create and track individual work items and tasks
        </li>
        <li>
          <strong>Teams:</strong> Organize your members into functional or project-based teams
        </li>
        <li>
          <strong>Settings:</strong> Configure your workspace, manage members, and customize preferences
        </li>
      </ul>

      <h2>Best Practices</h2>
      <ul>
        <li>Keep your workspace organized with clear naming conventions</li>
        <li>Regularly update progress on OKRs and stories</li>
        <li>Use labels and tags to categorize work effectively</li>
        <li>Encourage team collaboration through comments and updates</li>
        <li>Schedule regular check-ins to maintain alignment</li>
      </ul>`,
			"High",
		},
		{
			"Setting Up Your First OKR",
			"Learn how to create effective OKRs that align your team and drive measurable results.",
			`<h1>Creating Effective OKRs</h1>
      <p>OKRs (Objectives and Key Results) help align your team around measurable goals. This guide will help you create impactful OKRs that drive results.</p>

      <h2>Understanding OKRs</h2>
      <p>An OKR consists of two main components:</p>
      <ul>
        <li>
          <strong>Objective:</strong> A qualitative, inspirational goal that sets a clear direction
        </li>
        <li>
          <strong>Key Results:</strong> Quantitative metrics that measure progress toward the objective
        </li>
      </ul>

      <h2>Writing Objectives</h2>
      <p>Great objectives should be:</p>
      <ul>
        <li>Ambitious but achievable</li>
        <li>Clear and actionable</li>
        <li>Aligned with company goals</li>
        <li>Time-bound (usually quarterly)</li>
        <li>Inspiring and motivational</li>
      </ul>

      <h3>Example Objectives</h3>
      <ul>
        <li>✅ "Become the market leader in customer satisfaction"</li>
        <li>✅ "Transform our product into a must-have tool for developers"</li>
        <li>✅ "Build a world-class customer support experience"</li>
        <li>❌ "Increase revenue" (too vague)</li>
        <li>❌ "Improve the product" (not specific enough)</li>
      </ul>

      <h2>Defining Key Results</h2>
      <p>Effective key results should be:</p>
      <ul>
        <li>Measurable with specific numbers</li>
        <li>Challenging but realistic</li>
        <li>Directly tied to the objective</li>
        <li>Verifiable through data</li>
        <li>Time-bound within the objective's timeframe</li>
      </ul>

      <h3>Types of Key Results</h3>
      <ul>
        <li>
          <strong>Growth Metrics:</strong>
          <ul>
            <li>Revenue targets</li>
            <li>User acquisition goals</li>
            <li>Market share percentages</li>
          </ul>
        </li>
        <li>
          <strong>Performance Metrics:</strong>
          <ul>
            <li>Speed improvements</li>
            <li>Error rate reductions</li>
            <li>Efficiency gains</li>
          </ul>
        </li>
        <li>
          <strong>Quality Metrics:</strong>
          <ul>
            <li>Customer satisfaction scores</li>
            <li>User ratings</li>
            <li>SLA compliance rates</li>
          </ul>
        </li>
      </ul>

      <h2>Example OKRs</h2>
      
      <h3>Product Team Example</h3>
      <p><strong>Objective:</strong> Transform our product into the most user-friendly analytics platform</p>
      <p><strong>Key Results:</strong></p>
      <ul>
        <li>Increase user engagement rate from 60% to 85%</li>
        <li>Reduce time-to-insight from 5 minutes to 30 seconds</li>
        <li>Achieve a user satisfaction score of 9/10</li>
        <li>Decrease support tickets related to usability by 70%</li>
      </ul>

      <h3>Customer Success Team Example</h3>
      <p><strong>Objective:</strong> Become the industry benchmark for customer support excellence</p>
      <p><strong>Key Results:</strong></p>
      <ul>
        <li>Increase NPS score from 30 to 50</li>
        <li>Reduce first response time from 24h to 4h</li>
        <li>Achieve 95% customer satisfaction rating</li>
        <li>Increase customer retention rate from 85% to 95%</li>
      </ul>

      <h2>OKR Best Practices</h2>
      <ul>
        <li>
          <strong>Frequency and Timeline</strong>
          <ul>
            <li>Set company-level OKRs quarterly</li>
            <li>Review progress weekly or bi-weekly</li>
            <li>Adjust targets if necessary (but not too often)</li>
          </ul>
        </li>
        <li>
          <strong>Alignment</strong>
          <ul>
            <li>Ensure team OKRs support company objectives</li>
            <li>Create clear connections between different team OKRs</li>
            <li>Avoid conflicting objectives across teams</li>
          </ul>
        </li>
        <li>
          <strong>Tracking</strong>
          <ul>
            <li>Update progress regularly</li>
            <li>Use data to measure results</li>
            <li>Document learnings and adjustments</li>
          </ul>
        </li>
      </ul>

      <h2>Common Pitfalls to Avoid</h2>
      <ul>
        <li>Setting too many objectives (stick to 3-5 per team)</li>
        <li>Making key results without measurable outcomes</li>
        <li>Creating sandbagged or impossible targets</li>
        <li>Forgetting to align with company strategy</li>
        <li>Not reviewing and adjusting regularly</li>
      </ul>`,
			"High",
		},
	}

	var unstartedState states.CoreState
	for _, status := range statuses {
		if status.Category == "unstarted" {
			unstartedState = status
			break
		}
	}

	if unstartedState.ID == uuid.Nil {
		unstartedState = statuses[0]
	}

	s := make([]stories.CoreNewStory, len(storyData))
	for i, data := range storyData {
		desc := data.description
		descHTML := data.descriptionHTML
		s[i] = stories.CoreNewStory{
			Title:           data.title,
			Description:     &desc,
			DescriptionHTML: &descHTML,
			Reporter:        &userID,
			Assignee:        &userID,
			Priority:        data.priority,
			Team:            teamID,
			Status:          &unstartedState.ID,
		}
	}

	return s
}
