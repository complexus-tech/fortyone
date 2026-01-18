import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getRunningSprints } from "@/modules/sprints/queries/get-running-sprints";

export const listRunningSprints = tool({
  description:
    "List all currently running sprints. Returns active sprints with their details.",
  inputSchema: z.object({}),

  execute: async ({}, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access sprints",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const sprints = await getRunningSprints(ctx);

      return {
        success: true,
        sprints,
        count: sprints.length,
        message: `Found ${sprints.length} running sprint${sprints.length !== 1 ? "s" : ""}.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to list running sprints",
      };
    }
  },
});
