import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getTeamSettings } from "@/modules/teams/queries/get-team-settings";
import { getTeamMembersPage } from "@/lib/queries/members/get-members";

export const getTeamDetails = tool({
  description:
    "Get detailed information about a specific team including its members and settings.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to get details for (required)"),
    includeMembers: z
      .boolean()
      .optional()
      .describe("Include a first page of team members. Default true."),
    membersPageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of members to include when includeMembers is true."),
  }),

  execute: async (
    { teamId, includeMembers = true, membersPageSize = 20 },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access team details",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const [teams, settings, membersPage] = await Promise.all([
        getTeams(ctx),
        getTeamSettings(teamId, ctx).catch(() => null),
        includeMembers
          ? getTeamMembersPage(teamId, ctx, "", 1, membersPageSize)
          : Promise.resolve(null),
      ]);
      const team = teams.find((t) => t.id === teamId);

      if (!team) {
        return {
          success: false,
          error: "Team not found or you don't have access to it",
        };
      }

      return {
        success: true,
        team: {
          id: team.id,
          name: team.name,
          code: team.code,
          memberCount: team.memberCount,
          isPrivate: team.isPrivate,
          color: team.color,
        },
        settings,
        members: membersPage?.members,
        membersPagination: membersPage?.pagination,
        message: `Retrieved details for team "${team.name}".`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to get team details",
      };
    }
  },
});
