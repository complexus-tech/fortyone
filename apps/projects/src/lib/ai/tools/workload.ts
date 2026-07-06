import { tool } from "ai";
import { z } from "zod";
import { auth } from "@/auth";
import { getWorkloadAnalysis } from "@/modules/analytics/queries/get-workload-analysis";
import type { AnalyticsFilters } from "@/modules/analytics/types";

const compactFilters = (filters: AnalyticsFilters) => {
  return Object.fromEntries(
    Object.entries(filters).filter(([, value]) =>
      Array.isArray(value) ? value.length > 0 : Boolean(value),
    ),
  ) as AnalyticsFilters;
};

export const workloadPlanningTool = tool({
  description:
    "Answer workload and planning questions from the backend workload analysis report. Use for overloaded people, unassigned work, overdue/high-priority work, team workload, sprint workload, and who may need help.",
  inputSchema: z.object({
    teamIds: z
      .array(z.string())
      .optional()
      .describe(
        "Team IDs to scope workload. Omit for workspace-wide workload.",
      ),
    assigneeIds: z
      .array(z.string())
      .optional()
      .describe(
        "Optional user IDs to focus on after loading the workload report.",
      ),
    sprintIds: z.array(z.string()).optional().describe("Sprint filters."),
    objectiveIds: z.array(z.string()).optional().describe("Objective filters."),
    objectiveId: z
      .string()
      .optional()
      .describe("Single objective ID. Prefer objectiveIds for multiple."),
    startDate: z
      .string()
      .optional()
      .describe("Optional report start date, as ISO date or ISO datetime."),
    endDate: z
      .string()
      .optional()
      .describe("Optional report end date, as ISO date or ISO datetime."),
  }),
  execute: async (
    {
      teamIds,
      assigneeIds,
      sprintIds,
      objectiveIds,
      objectiveId,
      startDate,
      endDate,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to inspect workload",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug?: string })
        .workspaceSlug;

      if (!workspaceSlug) {
        return { success: false, error: "Workspace context is required" };
      }

      const filters = compactFilters({
        teamIds,
        assigneeIds,
        sprintIds,
        objectiveIds: objectiveIds ?? (objectiveId ? [objectiveId] : undefined),
        startDate,
        endDate,
      });
      const analysis = await getWorkloadAnalysis(
        { session, workspaceSlug },
        filters,
      );

      return {
        success: true,
        kind: "workload-analysis-report",
        filters,
        analysis,
        message: `Found workload for ${analysis.members.length} member${analysis.members.length === 1 ? "" : "s"} across ${analysis.teams.length} team${analysis.teams.length === 1 ? "" : "s"}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to inspect workload",
      };
    }
  },
});
