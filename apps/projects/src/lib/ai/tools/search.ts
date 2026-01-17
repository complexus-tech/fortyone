import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getMembers } from "@/lib/queries/members/get-members";
import { searchQuery } from "@/modules/search/queries/search";
import type { SearchQueryParams } from "@/modules/search/types";

export const searchTool = tool({
  description:
    "Search across all content types (stories, objectives) with advanced filtering and role-based permissions. Provides unified search across the entire workspace.",
  inputSchema: z.object({
    action: z
      .enum([
        "search-all",
        "search-stories",
        "search-objectives",
        "search-by-filters",
      ])
      .describe("The type of search to perform"),

    query: z
      .string()
      .optional()
      .describe("Search query to find content by title and description"),

    teamId: z.string().optional().describe("Team ID to filter results"),

    assigneeId: z
      .string()
      .optional()
      .describe("Assignee user ID to filter results"),

    statusId: z.string().optional().describe("Status ID to filter results"),

    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .optional()
      .describe("Filter by priority level"),

    sortBy: z
      .enum(["relevance", "updated", "created"])
      .optional()
      .describe("Sort order for results (default: relevance)"),

    page: z
      .number()
      .min(1)
      .optional()
      .describe("Page number for pagination (default: 1)"),

    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of results per page (default: 20, max: 100)"),

    includeDetails: z
      .boolean()
      .optional()
      .describe("Include detailed information in results (default: false)"),
  }),

  execute: async ({
    action,
    query,
    teamId,
    assigneeId,
    statusId,
    priority,
    sortBy = "relevance",
    page = 1,
    pageSize = 20,
    includeDetails = false,
  }) => {
    try {
      const session = await auth();

      if (!session) {


        return {
          success: false,
          error: "Authentication required to perform search",
        };
      }

      // Build search parameters based on action
      let searchParams: SearchQueryParams;

      switch (action) {
        case "search-all":
          searchParams = {
            type: "all",
            query,
            ...(teamId && { teamId }),
            ...(assigneeId && { assigneeId }),
            ...(statusId && { statusId }),
            ...(priority && { priority }),
            sortBy,
            page,
            pageSize,
          };
          break;

        case "search-stories":
          searchParams = {
            type: "stories",
            query,
            ...(teamId && { teamId }),
            ...(assigneeId && { assigneeId }),
            ...(statusId && { statusId }),
            ...(priority && { priority }),
            sortBy,
            page,
            pageSize,
          };
          break;

        case "search-objectives":
          searchParams = {
            type: "objectives",
            query,
            ...(teamId && { teamId }),
            ...(assigneeId && { assigneeId }),
            ...(statusId && { statusId }),
            sortBy,
            page,
            pageSize,
          };
          break;

        case "search-by-filters":
          searchParams = {
            type: "all",
            query,
            ...(teamId && { teamId }),
            ...(assigneeId && { assigneeId }),
            ...(statusId && { statusId }),
            ...(priority && { priority }),
            sortBy,
            page,
            pageSize,
          };
          break;

        default:
          return {
            success: false,
            error: "Invalid search action",
          };
      }

      // Perform the search
      const searchResults = await searchQuery(session, searchParams);

      // Get lookup data for enriching results
      const [allStatuses, allTeams, allMembers] = await Promise.all([
        getStatuses(session),
        getTeams(session),
        getMembers(session),
      ]);

      // Create lookup maps
      const statusMap = new Map(allStatuses.map((s) => [s.id, s]));
      const teamMap = new Map(allTeams.map((t) => [t.id, t]));
      const memberMap = new Map(allMembers.map((m) => [m.id, m]));

      // Format stories results
      const formattedStories = searchResults.stories.map((story) => {
        const status = statusMap.get(story.statusId);
        const team = teamMap.get(story.teamId);
        const assignee = story.assigneeId
          ? memberMap.get(story.assigneeId)
          : null;

        const baseStory = {
          id: story.id,
          title: story.title,
          priority: story.priority,
          statusId: story.statusId,
          status: status
            ? {
                name: status.name,
                color: status.color,
                category: status.category,
              }
            : null,
          assigneeId: story.assigneeId,
          assignee: assignee
            ? {
                id: assignee.id,
                fullName: assignee.fullName,
                username: assignee.username,
              }
            : null,
          teamId: story.teamId,
          team: team
            ? {
                id: team.id,
                name: team.name,
                code: team.code,
              }
            : null,
          createdAt: story.createdAt,
          updatedAt: story.updatedAt,
        };

        if (includeDetails) {
          return {
            ...baseStory,
            description: story.description,
            endDate: story.endDate,
            sprintId: story.sprintId,
            objectiveId: story.objectiveId,
          };
        }

        return baseStory;
      });

      // Format objectives results
      const formattedObjectives = searchResults.objectives.map((objective) => {
        const team = teamMap.get(objective.teamId);
        const leadUser = objective.leadUser
          ? memberMap.get(objective.leadUser)
          : null;

        const baseObjective = {
          id: objective.id,
          name: objective.name,
          teamId: objective.teamId,
          team: team
            ? {
                id: team.id,
                name: team.name,
                code: team.code,
              }
            : null,
          leadUser: leadUser
            ? {
                id: leadUser.id,
                fullName: leadUser.fullName,
                username: leadUser.username,
              }
            : null,
          startDate: objective.startDate,
          endDate: objective.endDate,
          priority: objective.priority,
          health: objective.health,
          createdAt: objective.createdAt,
          updatedAt: objective.updatedAt,
        };

        if (includeDetails) {
          return {
            ...baseObjective,
            description: objective.description,
            statusId: objective.statusId,
            isPrivate: objective.isPrivate,
            stats: objective.stats,
          };
        }

        return baseObjective;
      });

      // Build response message
      let message: string;
      const totalResults =
        searchResults.totalStories + searchResults.totalObjectives;

      if (query) {
        if (action === "search-stories") {
          message = `Search for "${query}" found ${searchResults.totalStories} stories.`;
        } else if (action === "search-objectives") {
          message = `Search for "${query}" found ${searchResults.totalObjectives} objectives.`;
        } else {
          message = `Search for "${query}" found ${totalResults} results (${searchResults.totalStories} stories, ${searchResults.totalObjectives} objectives).`;
        }
      } else if (action === "search-stories") {
        message = `Found ${searchResults.totalStories} stories.`;
      } else if (action === "search-objectives") {
        message = `Found ${searchResults.totalObjectives} objectives.`;
      } else {
        message = `Found ${totalResults} results (${searchResults.totalStories} stories, ${searchResults.totalObjectives} objectives).`;
      }

      return {
        success: true,
        query,
        action,
        stories: formattedStories,
        objectives: formattedObjectives,
        totals: {
          stories: searchResults.totalStories,
          objectives: searchResults.totalObjectives,
          total: totalResults,
        },
        pagination: {
          page: searchResults.page,
          pageSize: searchResults.pageSize,
          totalPages: searchResults.totalPages,
        },
        filters: {
          ...(teamId && { teamId }),
          ...(assigneeId && { assigneeId }),
          ...(statusId && { statusId }),
          ...(priority && { priority }),
        },
        message,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred during search",
      };
    }
  },
});
