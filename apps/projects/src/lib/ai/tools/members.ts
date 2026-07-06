import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import {
  getMembers,
  getMembersPage,
  getTeamMembersPage,
} from "@/lib/queries/members/get-members";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { resolvePaginationInput } from "./tool-helpers";

export const membersTool = tool({
  description:
    "Manage and view workspace members and team members based on user permissions",
  inputSchema: z.object({
    action: z
      .enum([
        "list-all-members",
        "list-team-members",
        "search-members",
        "get-member-details",
      ])
      .describe("The member operation to perform"),

    teamId: z
      .string()
      .optional()
      .describe("Team ID for team-specific member operations"),

    searchQuery: z
      .string()
      .optional()
      .describe("Search query to find members by name or username"),

    memberId: z
      .string()
      .optional()
      .describe("Member ID for specific member operations"),

    page: z.number().min(1).optional().describe("Page number. Default 1."),

    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of members per page. Default 20, max 100."),
  }),

  execute: async (
    { action, teamId, searchQuery, memberId, page, pageSize },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access members",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };
      const pagination = resolvePaginationInput({ page, pageSize });

      switch (action) {
        case "list-all-members": {
          const response = await getMembersPage(
            ctx,
            "",
            pagination.page,
            pagination.pageSize,
          );

          return {
            success: true,
            members: response.members.map((member) => ({
              id: member.id,
              name: member.fullName,
              username: member.username,
              role: member.role,
              avatarUrl: member.avatarUrl,
            })),
            pagination: response.pagination,
            count: response.members.length,
            message: `Found ${response.members.length} members in the workspace.`,
          };
        }

        case "list-team-members": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for listing team members",
            };
          }

          const [response, teams] = await Promise.all([
            getTeamMembersPage(
              teamId,
              ctx,
              "",
              pagination.page,
              pagination.pageSize,
            ),
            getTeams(ctx),
          ]);

          const team = teams.find((t) => t.id === teamId);
          const teamName = team?.name || "Unknown Team";

          return {
            success: true,
            teamId,
            teamName,
            members: response.members.map((member) => ({
              id: member.id,
              name: member.fullName,
              username: member.username,
              role: member.role,
              avatarUrl: member.avatarUrl,
            })),
            pagination: response.pagination,
            count: response.members.length,
            message: `Found ${response.members.length} members in ${teamName}.`,
          };
        }

        case "search-members": {
          if (!searchQuery) {
            return {
              success: false,
              error: "Search query is required for searching members",
            };
          }

          const response = await getMembersPage(
            ctx,
            searchQuery,
            pagination.page,
            pagination.pageSize,
          );

          return {
            success: true,
            members: response.members.map((member) => ({
              id: member.id,
              name: member.fullName,
              username: member.username,
              role: member.role,
              avatarUrl: member.avatarUrl,
            })),
            pagination: response.pagination,
            count: response.members.length,
            query: searchQuery,
            message: `Found ${response.members.length} members matching "${searchQuery}".`,
          };
        }

        case "get-member-details": {
          if (!memberId) {
            return {
              success: false,
              error: "Member ID is required for getting member details",
            };
          }

          const allMembers = await getMembers(ctx);
          const member = allMembers.find((m) => m.id === memberId);

          if (!member) {
            return {
              success: false,
              error: "Member not found or access denied",
            };
          }

          return {
            success: true,
            member: {
              id: member.id,
              name: member.fullName,
              username: member.username,
              role: member.role,
              avatarUrl: member.avatarUrl,
            },
            message: `Here are the details for ${member.fullName}.`,
          };
        }

        default:
          return {
            success: false,
            error: "Invalid action specified",
          };
      }
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred while performing member operation",
      };
    }
  },
});
