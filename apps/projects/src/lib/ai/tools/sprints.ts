import { z } from "zod";

export const sprintsTool = {
  name: "sprints",
  description:
    "Manage and query sprints, get sprint summaries and planning information",
  parameters: z.object({
    action: z
      .enum([
        "list-active",
        "list-all",
        "get-summary",
        "get-burndown",
        "get-velocity",
        "plan-next",
        "get-details",
      ])
      .describe("The action to perform on sprints"),
    sprintId: z
      .string()
      .optional()
      .describe("Sprint ID for specific sprint actions"),
    teamId: z.string().optional().describe("Team ID to filter sprints"),
    includeStories: z
      .boolean()
      .optional()
      .describe("Include story details in response"),
  }),
  execute: ({
    action,
    sprintId,
    teamId,
    includeStories,
  }: {
    action: string;
    sprintId?: string;
    teamId?: string;
    includeStories?: boolean;
  }) => {
    // This would integrate with your actual sprint management system
    // For now, returning mock data structure

    switch (action) {
      case "list-active":
        return {
          success: true,
          sprints: [
            {
              id: "sprint-1",
              name: "Sprint 15 - User Authentication",
              team: "Frontend Team",
              startDate: "2024-02-01",
              endDate: "2024-02-14",
              status: "active",
              progress: 65,
              totalPoints: 21,
              completedPoints: 14,
            },
            {
              id: "sprint-2",
              name: "Sprint 16 - Dashboard Features",
              team: "Backend Team",
              startDate: "2024-02-01",
              endDate: "2024-02-14",
              status: "active",
              progress: 45,
              totalPoints: 18,
              completedPoints: 8,
            },
          ],
          count: 2,
          message: "Here are the currently active sprints.",
        };

      case "list-all":
        return {
          success: true,
          sprints: [
            {
              id: "sprint-1",
              name: "Sprint 15 - User Authentication",
              team: "Frontend Team",
              startDate: "2024-02-01",
              endDate: "2024-02-14",
              status: "active",
              progress: 65,
            },
            {
              id: "sprint-0",
              name: "Sprint 14 - Initial Setup",
              team: "Frontend Team",
              startDate: "2024-01-15",
              endDate: "2024-01-31",
              status: "completed",
              progress: 100,
            },
          ],
          count: 2,
          message: "Here are all sprints for the current view.",
        };

      case "get-summary": {
        const targetSprintId = sprintId || "sprint-1";
        return {
          success: true,
          sprint: {
            id: targetSprintId,
            name: "Sprint 15 - User Authentication",
            team: "Frontend Team",
            startDate: "2024-02-01",
            endDate: "2024-02-14",
            status: "active",
            progress: 65,
            totalPoints: 21,
            completedPoints: 14,
            remainingPoints: 7,
            daysRemaining: 3,
            velocity: 2.3,
            stories: includeStories
              ? [
                  {
                    id: "story-1",
                    title: "Implement login form",
                    status: "completed",
                    points: 5,
                    assignee: "John Doe",
                  },
                  {
                    id: "story-2",
                    title: "Add password reset functionality",
                    status: "in-progress",
                    points: 8,
                    assignee: "Jane Smith",
                  },
                  {
                    id: "story-3",
                    title: "Setup OAuth integration",
                    status: "todo",
                    points: 8,
                    assignee: "Mike Johnson",
                  },
                ]
              : undefined,
          },
          message: `Here's the summary for ${targetSprintId}.`,
        };
      }

      case "get-burndown":
        return {
          success: true,
          burndown: {
            sprintId: sprintId || "sprint-1",
            data: [
              { day: 1, remainingPoints: 21 },
              { day: 2, remainingPoints: 19 },
              { day: 3, remainingPoints: 18 },
              { day: 4, remainingPoints: 16 },
              { day: 5, remainingPoints: 14 },
              { day: 6, remainingPoints: 12 },
              { day: 7, remainingPoints: 10 },
              { day: 8, remainingPoints: 8 },
              { day: 9, remainingPoints: 7 },
              { day: 10, remainingPoints: 7 },
            ],
            idealLine: [
              { day: 1, remainingPoints: 21 },
              { day: 10, remainingPoints: 0 },
            ],
          },
          message: "Here is the burndown chart data for the sprint.",
        };

      case "get-velocity":
        return {
          success: true,
          velocity: {
            team: teamId || "Frontend Team",
            lastSprints: [
              { sprintId: "sprint-14", name: "Sprint 14", velocity: 2.1 },
              { sprintId: "sprint-13", name: "Sprint 13", velocity: 2.4 },
              { sprintId: "sprint-12", name: "Sprint 12", velocity: 2.0 },
              { sprintId: "sprint-11", name: "Sprint 11", velocity: 2.3 },
            ],
            averageVelocity: 2.2,
            trend: "stable",
          },
          message: "Here is the velocity data for the team.",
        };

      case "plan-next":
        return {
          success: true,
          planning: {
            suggestedSprintName: "Sprint 16 - Dashboard Enhancements",
            suggestedDuration: "2 weeks",
            availableStories: [
              {
                id: "story-4",
                title: "Add analytics dashboard",
                points: 13,
                priority: "high",
              },
              {
                id: "story-5",
                title: "Implement data export",
                points: 8,
                priority: "medium",
              },
              {
                id: "story-6",
                title: "Add user preferences",
                points: 5,
                priority: "low",
              },
            ],
            recommendedStories: ["story-4", "story-5"],
            estimatedVelocity: 2.2,
            recommendedPoints: 21,
          },
          message: "Here are my recommendations for the next sprint planning.",
        };

      case "get-details":
        if (!sprintId) {
          return {
            success: false,
            error: "Sprint ID is required for get-details action",
          };
        }
        return {
          success: true,
          sprint: {
            id: sprintId,
            name: "Sprint 15 - User Authentication",
            team: "Frontend Team",
            startDate: "2024-02-01",
            endDate: "2024-02-14",
            status: "active",
            progress: 65,
            totalPoints: 21,
            completedPoints: 14,
            remainingPoints: 7,
            daysRemaining: 3,
            velocity: 2.3,
            goals: [
              "Complete user authentication flow",
              "Implement password reset functionality",
              "Setup OAuth integration",
            ],
            risks: [
              "OAuth provider integration complexity",
              "Security review requirements",
            ],
          },
          message: `Here are the detailed information for sprint ${sprintId}.`,
        };

      default:
        return {
          success: false,
          error: `Unknown action: ${action}`,
        };
    }
  },
};
