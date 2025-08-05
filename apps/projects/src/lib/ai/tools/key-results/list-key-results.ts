import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getWorkspaceKeyResults } from "@/modules/key-results/queries/get-workspace-key-results";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const listKeyResultsTool = tool({
  description:
    "List and filter key results across the workspace with team-based permissions and flexible filtering options.",
  parameters: z.object({
    filters: z
      .object({
        objectiveIds: z.array(z.string()).optional(),
        teamIds: z.array(z.string()).optional(),
        measurementTypes: z
          .array(z.enum(["percentage", "number", "boolean"]))
          .optional(),
        createdAfter: z.string().optional(),
        createdBefore: z.string().optional(),
        updatedAfter: z.string().optional(),
        updatedBefore: z.string().optional(),
        page: z.number().min(1).optional(),
        pageSize: z.number().min(1).max(100).optional(),
        orderBy: z
          .enum(["name", "created_at", "updated_at", "objective_name"])
          .optional(),
        orderDirection: z.enum(["asc", "desc"]).optional(),
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
        lastUpdatedBy: kr.lastUpdatedBy,
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
