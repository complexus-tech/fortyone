import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeamMembersPage } from "@/lib/queries/members/get-members";
import { resolvePaginationInput } from "../tool-helpers";

export const listTeamMembers = tool({
  description:
    "List all members of a specific team. Returns member details including roles and contact information.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to list members for (required)"),
    searchQuery: z
      .string()
      .optional()
      .describe("Optional search query for member name or username."),
    page: z.number().min(1).optional().describe("Page number. Default 1."),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of team members per page. Default 20, max 100."),
  }),

  execute: async (
    { teamId, searchQuery, page, pageSize },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access team members",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const pagination = resolvePaginationInput({ page, pageSize });
      const response = await getTeamMembersPage(
        teamId,
        ctx,
        searchQuery,
        pagination.page,
        pagination.pageSize,
      );

      return {
        success: true,
        members: response.members.map((member) => ({
          id: member.id,
          fullName: member.fullName,
          username: member.username,
          email: member.email,
          avatarUrl: member.avatarUrl,
          role: member.role,
        })),
        pagination: response.pagination,
        message: `Found ${response.members.length} member${response.members.length !== 1 ? "s" : ""} in this team.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to list team members",
      };
    }
  },
});
