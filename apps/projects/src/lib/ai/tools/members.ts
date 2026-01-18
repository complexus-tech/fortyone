import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getMembers, getTeamMembers } from "@/lib/queries/members/get-members";
import { getTeams } from "@/modules/teams/queries/get-teams";

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
  }),

  execute: async ({ action, teamId, searchQuery, memberId }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access members",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      switch (action) {
        case "list-all-members": {
          const members = await getMembers(ctx);

          return {
            success: true,
            members: members.map((member) => ({
              id: member.id,
              name: member.fullName,
              username: member.username,
              role: member.role,
              avatarUrl: member.avatarUrl,
            })),
            count: members.length,
            message: `Found ${members.length} members in the workspace.`,
          };
        }

        case "list-team-members": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for listing team members",
            };
          }

          const [members, teams] = await Promise.all([
            getTeamMembers(teamId, ctx),
            getTeams(ctx),
          ]);

          const team = teams.find((t) => t.id === teamId);
          const teamName = team?.name || "Unknown Team";

          return {
            success: true,
            teamId,
            teamName,
            members: members.map((member) => ({
              id: member.id,
              name: member.fullName,
              username: member.username,
              role: member.role,
              avatarUrl: member.avatarUrl,
            })),
            count: members.length,
            message: `Found ${members.length} members in ${teamName}.`,
          };
        }

        case "search-members": {
          if (!searchQuery) {
            return {
              success: false,
              error: "Search query is required for searching members",
            };
          }

          const allMembers = await getMembers(ctx);
          const query = searchQuery.toLowerCase();

          const matchingMembers = allMembers.filter(
            (member) =>
              member.fullName.toLowerCase().includes(query) ||
              member.username.toLowerCase().includes(query),
          );

          return {
            success: true,
            members: matchingMembers.map((member) => ({
              id: member.id,
              name: member.fullName,
              username: member.username,
              role: member.role,
              avatarUrl: member.avatarUrl,
            })),
            count: matchingMembers.length,
            query: searchQuery,
            message: `Found ${matchingMembers.length} members matching "${searchQuery}".`,
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
