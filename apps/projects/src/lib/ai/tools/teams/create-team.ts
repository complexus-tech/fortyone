import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createTeam } from "@/modules/teams/actions";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const createTeamTool = tool({
  description:
    "Create a new team. Only admins and members can create teams. Guests cannot create teams.",
  inputSchema: z.object({
    name: z.string().describe("Team name (required)"),
    code: z
      .string()
      .describe("Team code (unique identifier) (required) (3 characters)"),
    color: z
      .string()
      .describe(
        "Team color (hex code) (required) - suggest colors that work in both dark and light mode",
      ),
    isPrivate: z
      .boolean()
      .optional()
      .describe("Whether team is private (default: false)"),
  }),

  execute: async ({ name, code, color, isPrivate = false }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to create teams",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      if (userRole === "guest") {
        return {
          success: false,
          error: "Guests cannot create teams",
        };
      }

      const teamData = {
        name,
        code,
        color,
        isPrivate,
      };

      const result = await createTeam(teamData, workspaceSlug);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to create team",
        };
      }

      return {
        success: true,
        team: result.data,
        message: `Team "${name}" created successfully.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to create team",
      };
    }
  },
});
