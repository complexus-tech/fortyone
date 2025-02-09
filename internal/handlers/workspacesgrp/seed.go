package workspacesgrp

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
		{
			"Sprint Planning Guide",
			"Master the art of sprint planning and execution to deliver value consistently.",
			`<h1>Sprint Planning Guide</h1>
      <p>Effective sprint planning helps your team deliver value consistently. This comprehensive guide will help you master the sprint process from start to finish.</p>

      <h2>Sprint Fundamentals</h2>
      <p>A sprint is a fixed-length iteration of work that follows these principles:</p>
      <ul>
        <li>Fixed timeframe (typically 1-4 weeks)</li>
        <li>Clear, achievable goals</li>
        <li>Focused team effort</li>
        <li>Deliverable increments of value</li>
      </ul>

      <h2>Sprint Structure</h2>
      <h3>1. Sprint Planning</h3>
      <ul>
        <li>
          <strong>Sprint Goal Setting</strong>
          <ul>
            <li>Define clear objectives for the sprint</li>
            <li>Align with broader project/product goals</li>
            <li>Ensure the goal is achievable within the sprint</li>
          </ul>
        </li>
        <li>
          <strong>Capacity Planning</strong>
          <ul>
            <li>Calculate available team hours</li>
            <li>Account for meetings and other commitments</li>
            <li>Consider team velocity from previous sprints</li>
            <li>Factor in planned time off or holidays</li>
          </ul>
        </li>
        <li>
          <strong>Story Selection</strong>
          <ul>
            <li>Choose stories that align with sprint goal</li>
            <li>Ensure stories are properly sized</li>
            <li>Confirm all prerequisites are met</li>
            <li>Include a mix of story types (features, bugs, tech debt)</li>
          </ul>
        </li>
      </ul>

      <h3>2. Daily Execution</h3>
      <ul>
        <li>
          <strong>Daily Standup</strong>
          <ul>
            <li>Share progress (What did you do yesterday?)</li>
            <li>Plan for today (What will you do today?)</li>
            <li>Identify blockers (Any impediments?)</li>
            <li>Keep it brief (15 minutes max)</li>
          </ul>
        </li>
        <li>
          <strong>Progress Tracking</strong>
          <ul>
            <li>Update story status regularly</li>
            <li>Monitor burndown chart</li>
            <li>Address blockers promptly</li>
            <li>Communicate changes or challenges</li>
          </ul>
        </li>
      </ul>

      <h3>3. Sprint Review</h3>
      <ul>
        <li>
          <strong>Demo Preparation</strong>
          <ul>
            <li>Prepare working demonstrations</li>
            <li>Focus on value delivered</li>
            <li>Include relevant stakeholders</li>
          </ul>
        </li>
        <li>
          <strong>Feedback Collection</strong>
          <ul>
            <li>Gather stakeholder input</li>
            <li>Document feature requests</li>
            <li>Note areas for improvement</li>
          </ul>
        </li>
      </ul>

      <h3>4. Sprint Retrospective</h3>
      <ul>
        <li>
          <strong>Team Discussion</strong>
          <ul>
            <li>What went well?</li>
            <li>What could be improved?</li>
            <li>What actions should we take?</li>
          </ul>
        </li>
        <li>
          <strong>Action Items</strong>
          <ul>
            <li>Document agreed improvements</li>
            <li>Assign owners to action items</li>
            <li>Follow up on previous retro items</li>
          </ul>
        </li>
      </ul>

      <h2>Sprint Best Practices</h2>
      <ul>
        <li>
          <strong>Planning</strong>
          <ul>
            <li>Don't overcommit - leave room for unknowns</li>
            <li>Ensure stories are properly groomed</li>
            <li>Have clear acceptance criteria</li>
            <li>Include buffer for bugs and technical debt</li>
          </ul>
        </li>
        <li>
          <strong>Execution</strong>
          <ul>
            <li>Start with highest priority items</li>
            <li>Limit work in progress</li>
            <li>Address blockers immediately</li>
            <li>Maintain regular communication</li>
          </ul>
        </li>
        <li>
          <strong>Review and Improvement</strong>
          <ul>
            <li>Document lessons learned</li>
            <li>Track team velocity</li>
            <li>Refine estimation accuracy</li>
            <li>Continuously improve processes</li>
          </ul>
        </li>
      </ul>

      <h2>Common Sprint Challenges</h2>
      <ul>
        <li>
          <strong>Scope Creep</strong>
          <p>Solution: Strictly manage sprint boundaries and have a clear process for handling new requests</p>
        </li>
        <li>
          <strong>Incomplete Stories</strong>
          <p>Solution: Break down stories into smaller pieces and improve estimation accuracy</p>
        </li>
        <li>
          <strong>Blockers</strong>
          <p>Solution: Identify dependencies early and establish quick escalation paths</p>
        </li>
        <li>
          <strong>Team Availability</strong>
          <p>Solution: Account for all known absences during sprint planning and maintain a buffer</p>
        </li>
      </ul>`,
			"Medium",
		},
		{
			"Story Management",
			"Learn how to create, organize, and track stories effectively to manage your team's work.",
			`<h1>Managing Stories</h1>
      <p>Stories are the fundamental unit of work in agile development. This guide will help you create and manage effective stories that drive value delivery.</p>

      <h2>Story Anatomy</h2>
      <h3>Essential Components</h3>
      <ul>
        <li>
          <strong>Title:</strong>
          <ul>
            <li>Clear and concise description</li>
            <li>Unique identifier</li>
            <li>Searchable keywords</li>
          </ul>
        </li>
        <li>
          <strong>Description:</strong>
          <ul>
            <li>Detailed requirements</li>
            <li>Context and background</li>
            <li>User benefit or business value</li>
          </ul>
        </li>
        <li>
          <strong>Acceptance Criteria:</strong>
          <ul>
            <li>Specific conditions to be met</li>
            <li>Testable requirements</li>
            <li>Definition of "done"</li>
          </ul>
        </li>
        <li>
          <strong>Story Points:</strong>
          <ul>
            <li>Relative effort estimation</li>
            <li>Complexity consideration</li>
            <li>Risk assessment</li>
          </ul>
        </li>
      </ul>

      <h2>Story Writing Format</h2>
      <h3>User Story Template</h3>
      <pre>
As a [type of user]
I want to [perform some action]
So that [achieve some goal/value]
      </pre>

      <h3>Example Stories</h3>
      <h4>Feature Story</h4>
      <pre>
Title: User Password Reset

As a registered user
I want to reset my password
So that I can regain access to my account if I forget it

Acceptance Criteria:
1. User can request password reset from login page
2. Reset link is sent to registered email
3. Link expires after 24 hours
4. Password must meet security requirements
5. User is notified on successful reset

Story Points: 5
      </pre>

      <h4>Bug Story</h4>
      <pre>
Title: Mobile Navigation Menu Not Closing

As a mobile user
I want the navigation menu to close when I select an item
So that I can see the content I've navigated to

Acceptance Criteria:
1. Menu closes automatically after item selection
2. No console errors
3. Works on all supported mobile browsers
4. Animation is smooth

Story Points: 3
      </pre>

      <h2>Story States</h2>
      <p>Stories move through different states as they progress:</p>
      <ol>
        <li>
          <strong>Backlog</strong>
          <ul>
            <li>Initial state for new stories</li>
            <li>Needs refinement and prioritization</li>
            <li>May lack complete details</li>
          </ul>
        </li>
        <li>
          <strong>Ready</strong>
          <ul>
            <li>Fully detailed and understood</li>
            <li>Acceptance criteria defined</li>
            <li>Dependencies identified</li>
            <li>Sized and prioritized</li>
          </ul>
        </li>
        <li>
          <strong>In Progress</strong>
          <ul>
            <li>Currently being worked on</li>
            <li>Assigned to team member</li>
            <li>Regular status updates</li>
          </ul>
        </li>
        <li>
          <strong>Review</strong>
          <ul>
            <li>Work completed</li>
            <li>Ready for testing/review</li>
            <li>Pending approval</li>
          </ul>
        </li>
        <li>
          <strong>Done</strong>
          <ul>
            <li>Meets acceptance criteria</li>
            <li>Approved by stakeholders</li>
            <li>Ready for release</li>
          </ul>
        </li>
      </ol>

      <h2>Story Management Best Practices</h2>
      <h3>Writing Stories</h3>
      <ul>
        <li>Keep stories small and focused</li>
        <li>Use clear, unambiguous language</li>
        <li>Include all necessary context</li>
        <li>Link to relevant documentation</li>
        <li>Add visual aids when helpful (mockups, diagrams)</li>
      </ul>

      <h3>Story Refinement</h3>
      <ul>
        <li>
          <strong>Regular Grooming</strong>
          <ul>
            <li>Review and update stories regularly</li>
            <li>Break down large stories</li>
            <li>Clarify requirements</li>
            <li>Update priorities</li>
          </ul>
        </li>
        <li>
          <strong>Estimation</strong>
          <ul>
            <li>Use consistent point scale</li>
            <li>Consider all aspects of work</li>
            <li>Include team discussion</li>
            <li>Account for unknowns</li>
          </ul>
        </li>
      </ul>

      <h3>Story Organization</h3>
      <ul>
        <li>
          <strong>Labels and Tags</strong>
          <ul>
            <li>Use consistent labeling system</li>
            <li>Tag by type (feature, bug, tech debt)</li>
            <li>Mark priority levels</li>
            <li>Indicate components affected</li>
          </ul>
        </li>
        <li>
          <strong>Dependencies</strong>
          <ul>
            <li>Clearly mark blocked/blocking stories</li>
            <li>Link related stories</li>
            <li>Document external dependencies</li>
          </ul>
        </li>
      </ul>

      <h2>Story Lifecycle Management</h2>
      <ul>
        <li>
          <strong>Creation</strong>
          <ul>
            <li>Capture initial requirements</li>
            <li>Add basic details</li>
            <li>Set preliminary priority</li>
          </ul>
        </li>
        <li>
          <strong>Refinement</strong>
          <ul>
            <li>Add detailed requirements</li>
            <li>Define acceptance criteria</li>
            <li>Estimate effort</li>
          </ul>
        </li>
        <li>
          <strong>Implementation</strong>
          <ul>
            <li>Regular status updates</li>
            <li>Document decisions</li>
            <li>Track time spent</li>
          </ul>
        </li>
        <li>
          <strong>Review</strong>
          <ul>
            <li>Verify acceptance criteria</li>
            <li>Document test results</li>
            <li>Collect feedback</li>
          </ul>
        </li>
        <li>
          <strong>Closure</strong>
          <ul>
            <li>Confirm completion</li>
            <li>Document outcomes</li>
            <li>Capture lessons learned</li>
          </ul>
        </li>
      </ul>`,
			"Medium",
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
