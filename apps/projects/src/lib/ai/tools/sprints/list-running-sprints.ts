import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getRunningSprints } from "@/modules/sprints/queries/get-running-sprints";
import { paginateRecords } from "../tool-helpers";

export const listRunningSprints = tool({
  description:
    "List all currently running sprints. Returns active sprints with their details.",
  inputSchema: z.object({
    page: z.number().min(1).optional().describe("Page number. Default 1."),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of running sprints per page. Default 20, max 100."),
  }),

  execute: async (
    { page, pageSize },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access sprints",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const sprints = await getRunningSprints(ctx);
      const result = paginateRecords(sprints, { page, pageSize });

      return {
        success: true,
        sprints: result.records,
        count: result.records.length,
        pagination: result.pagination,
        message: `Found ${result.records.length} running sprint${result.records.length !== 1 ? "s" : ""}.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to list running sprints",
      };
    }
  },
});
