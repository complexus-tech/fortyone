import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeamMembers } from "@/lib/queries/members/get-members";

export const listTeamMembers = tool({
  description:
    "List all members of a specific team. Returns member details including roles and contact information.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to list members for (required)"),
  }),

  execute: async (({ teamId }), { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {


        return {
      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };
          success: false,
          error: "Authentication required to access team members",
        };
      }

      const members = await getTeamMembers(teamId, session);

      return {
        success: true,
        members: members.map((member) => ({
          id: member.id,
          fullName: member.fullName,
          username: member.username,
          email: member.email,
          avatarUrl: member.avatarUrl,
          role: member.role,
        })),
        message: `Found ${members.length} member${members.length !== 1 ? "s" : ""} in this team.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to list team members",
      };
    }
  },
});
