import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getWorkspaceKeyResults } from "@/modules/key-results/queries/get-workspace-key-results";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const listKeyResultsTool = tool({
  description:
    "List and filter key results across the workspace with team-based permissions and flexible filtering options.",
  inputSchema: z.object({
    filters: z
      .object({
        objectiveIds: z
          .array(z.string())
          .optional()
          .describe("Filter key results by objective IDs"),
        teamIds: z
          .array(z.string())
          .optional()
          .describe("Filter key results by team IDs"),
        measurementTypes: z
          .array(z.enum(["percentage", "number", "boolean"]))
          .optional()
          .describe("Filter key results by measurement type"),
        createdAfter: z
          .string()
          .optional()
          .describe("Filter key results created after this date (ISO date)"),
        createdBefore: z
          .string()
          .optional()
          .describe("Filter key results created before this date (ISO date)"),
        updatedAfter: z
          .string()
          .optional()
          .describe("Filter key results updated after this date (ISO date)"),
        updatedBefore: z.string().optional(),
        page: z
          .number()
          .min(1)
          .optional()
          .describe("Page number for pagination"),
        pageSize: z
          .number()
          .min(1)
          .max(100)
          .optional()
          .describe("Number of key results per page"),
        orderBy: z
          .enum(["name", "created_at", "updated_at", "objective_name"])
          .optional()
          .describe("Field to order by"),
        orderDirection: z
          .enum(["asc", "desc"])
          .optional()
          .describe("Direction to order by"),
      })
      .optional()
      .describe("Filter options for listing key results"),
  }),

  execute: async ({ filters }) => {
    const session = await auth();

    if (!session) {
      return {
        success: false,
        error: "Authentication required",
      };
    }

    // Get user's workspace and role for permissions
    const workspace = await getWorkspace(session);
    const userRole = workspace.userRole;

    if (userRole === "guest") {
      return {
        success: false,
        error: "Guests cannot view key results",
      };
    }

    const response = await getWorkspaceKeyResults(session, filters);

    return {
      success: true,
      keyResults: response.keyResults.map((kr) => ({
        id: kr.id,
        name: kr.name,
        measurementType: kr.measurementType,
        startValue: kr.startValue,
        currentValue: kr.currentValue,
        targetValue: kr.targetValue,
        objectiveId: kr.objectiveId,
        objectiveName: kr.objectiveName,
        teamId: kr.teamId,
        teamName: kr.teamName,
        workspaceId: kr.workspaceId,
        createdAt: kr.createdAt,
        updatedAt: kr.updatedAt,
        createdBy: kr.createdBy,
        lead: kr.lead,
        contributors: kr.contributors,
        startDate: kr.startDate,
        endDate: kr.endDate,
      })),
      count: response.totalCount,
      pagination: {
        page: response.page,
        pageSize: response.pageSize,
        hasMore: response.hasMore,
      },
      message: `Found ${response.keyResults.length} key results.`,
    };
  },
});
