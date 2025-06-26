import { z } from "zod";

export const storiesTool = {
  name: "stories",
  description: "Manage and query stories/tasks in the application",
  parameters: z.object({
    action: z
      .enum([
        "list-assigned",
        "list-created",
        "list-all",
        "get-details",
        "create",
        "search",
      ])
      .describe("The action to perform on stories"),
    filters: z
      .object({
        status: z.string().optional().describe("Filter by status"),
        priority: z.string().optional().describe("Filter by priority"),
        assignee: z.string().optional().describe("Filter by assignee"),
        team: z.string().optional().describe("Filter by team"),
      })
      .optional()
      .describe("Optional filters to apply"),
    storyId: z.string().optional().describe("Story ID for get-details action"),
    searchQuery: z
      .string()
      .optional()
      .describe("Search query for search action"),
    storyData: z
      .object({
        title: z.string().describe("Story title"),
        description: z.string().optional().describe("Story description"),
        priority: z
          .enum(["low", "medium", "high", "urgent"])
          .optional()
          .describe("Story priority"),
        assignee: z.string().optional().describe("Assignee ID"),
        team: z.string().optional().describe("Team ID"),
        status: z.string().optional().describe("Initial status"),
      })
      .optional()
      .describe("Story data for create action"),
  }),
  execute: async ({
    action,
    filters,
    storyId,
    searchQuery,
    storyData,
  }: {
    action: string;
    filters?: any;
    storyId?: string;
    searchQuery?: string;
    storyData?: any;
  }) => {
    // This would integrate with your actual story management system
    // For now, returning mock data structure

    switch (action) {
      case "list-assigned":
        return {
          success: true,
          stories: [
            {
              id: "1",
              title: "Implement user authentication",
              status: "in-progress",
              priority: "high",
              assignee: "current-user",
              dueDate: "2024-02-15",
            },
            {
              id: "2",
              title: "Design dashboard layout",
              status: "todo",
              priority: "medium",
              assignee: "current-user",
              dueDate: "2024-02-20",
            },
          ],
          count: 2,
          message: "Here are the stories assigned to you.",
        };

      case "list-created":
        return {
          success: true,
          stories: [
            {
              id: "3",
              title: "Setup CI/CD pipeline",
              status: "completed",
              priority: "high",
              creator: "current-user",
              completedDate: "2024-02-10",
            },
          ],
          count: 1,
          message: "Here are the stories you created.",
        };

      case "list-all":
        return {
          success: true,
          stories: [
            {
              id: "1",
              title: "Implement user authentication",
              status: "in-progress",
              priority: "high",
              assignee: "current-user",
            },
            {
              id: "2",
              title: "Design dashboard layout",
              status: "todo",
              priority: "medium",
              assignee: "current-user",
            },
            {
              id: "3",
              title: "Setup CI/CD pipeline",
              status: "completed",
              priority: "high",
              creator: "current-user",
            },
          ],
          count: 3,
          message: "Here are all the stories in the current view.",
        };

      case "get-details":
        if (!storyId) {
          return {
            success: false,
            error: "Story ID is required for get-details action",
          };
        }
        return {
          success: true,
          story: {
            id: storyId,
            title: "Sample Story",
            description: "This is a detailed description of the story",
            status: "in-progress",
            priority: "high",
            assignee: "current-user",
            createdDate: "2024-02-01",
            updatedDate: "2024-02-12",
            comments: [
              {
                id: "1",
                author: "John Doe",
                content: "Started working on this",
                timestamp: "2024-02-12T10:00:00Z",
              },
            ],
          },
          message: `Here are the details for story ${storyId}.`,
        };

      case "create":
        if (!storyData) {
          return {
            success: false,
            error: "Story data is required for create action",
          };
        }
        return {
          success: true,
          story: {
            id: "new-story-id",
            ...storyData,
            createdDate: new Date().toISOString(),
            status: storyData.status || "todo",
          },
          message: `Successfully created story: ${storyData.title}`,
        };

      case "search":
        if (!searchQuery) {
          return {
            success: false,
            error: "Search query is required for search action",
          };
        }
        return {
          success: true,
          stories: [
            {
              id: "1",
              title: "Implement user authentication",
              status: "in-progress",
              priority: "high",
              assignee: "current-user",
            },
          ],
          count: 1,
          query: searchQuery,
          message: `Found 1 story matching "${searchQuery}".`,
        };

      default:
        return {
          success: false,
          error: `Unknown action: ${action}`,
        };
    }
  },
};
