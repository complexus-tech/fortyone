import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getRunningSprints } from "@/modules/sprints/queries/get-running-sprints";

export const listRunningSprints = tool({
  description:
    "List all currently running sprints. Returns active sprints with their details.",
  parameters: z.object({}),

  execute: async () => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access sprints",
        };
      }

      const sprints = await getRunningSprints(session);

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
