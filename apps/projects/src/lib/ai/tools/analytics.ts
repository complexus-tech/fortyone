import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import type { WorkspaceCtx } from "@/lib/http";
import { getWorkspaceOverview } from "@/modules/analytics/queries/get-workspace-overview";
import { getStoryAnalytics } from "@/modules/analytics/queries/get-story-analytics";
import { getObjectiveProgress } from "@/modules/analytics/queries/get-objective-progress";
import { getTeamPerformance } from "@/modules/analytics/queries/get-team-performance";
import { getSprintAnalytics } from "@/modules/analytics/queries/get-sprint-analytics";
import { getTimelineTrends } from "@/modules/analytics/queries/get-timeline-trends";
import type { AnalyticsFilters } from "@/modules/analytics/types";

const analyticsFiltersSchema = z.object({
  teamIds: z
    .array(z.string())
    .optional()
    .describe("Team IDs to filter by. Use listTeams first when needed."),
  startDate: z
    .string()
    .optional()
    .describe("Start date as an ISO date string, for example 2026-06-01."),
  endDate: z
    .string()
    .optional()
    .describe("End date as an ISO date string, for example 2026-06-30."),
  sprintIds: z
    .array(z.string())
    .optional()
    .describe("Sprint IDs to filter by. Use listSprints first when needed."),
  objectiveIds: z
    .array(z.string())
    .optional()
    .describe(
      "Objective IDs to filter by. Use listObjectivesTool first when needed.",
    ),
});

const compactFilters = (filters?: AnalyticsFilters) => {
  if (!filters) return undefined;

  return Object.fromEntries(
    Object.entries(filters).filter(([, value]) =>
      Array.isArray(value) ? value.length > 0 : Boolean(value),
    ),
  ) as AnalyticsFilters;
};

const getAuthenticatedContext = async (
  experimentalContext: unknown,
): Promise<WorkspaceCtx | { error: string }> => {
  const session = await auth();

  if (!session) {
    return { error: "Authentication required to access analytics" };
  }

  const workspaceSlug = (experimentalContext as { workspaceSlug?: string })
    .workspaceSlug;

  if (!workspaceSlug) {
    return { error: "Workspace context is required to access analytics" };
  }

  return { session, workspaceSlug };
};

export const workspacePerformanceReportTool = tool({
  description:
    "Build a workspace-level performance report with total stories, completed stories, active objectives, active sprints, members, completion trend, and velocity trend.",
  inputSchema: z.object({
    filters: analyticsFiltersSchema.optional(),
  }),
  execute: async ({ filters }, { experimental_context }) => {
    try {
      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const overview = await getWorkspaceOverview(ctx, compactFilters(filters));

      return {
        success: true,
        kind: "workspace-performance-report",
        title: "Workspace performance",
        filters: compactFilters(filters),
        overview,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to build workspace performance report",
      };
    }
  },
});

export const storyPerformanceReportTool = tool({
  description:
    "Build a story performance report with status breakdown, priority distribution, team completion, and burndown data.",
  inputSchema: z.object({
    filters: analyticsFiltersSchema.optional(),
  }),
  execute: async ({ filters }, { experimental_context }) => {
    try {
      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const analytics = await getStoryAnalytics(ctx, compactFilters(filters));

      return {
        success: true,
        kind: "story-performance-report",
        title: "Story performance",
        filters: compactFilters(filters),
        analytics,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to build story performance report",
      };
    }
  },
});

export const objectiveProgressReportTool = tool({
  description:
    "Build an objective progress report with objective health, status breakdown, key-result progress, and progress by team.",
  inputSchema: z.object({
    filters: analyticsFiltersSchema.optional(),
  }),
  execute: async ({ filters }, { experimental_context }) => {
    try {
      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const progress = await getObjectiveProgress(ctx, compactFilters(filters));

      return {
        success: true,
        kind: "objective-progress-report",
        title: "Objective progress",
        filters: compactFilters(filters),
        progress,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to build objective progress report",
      };
    }
  },
});

export const teamPerformanceReportTool = tool({
  description:
    "Build a team or person performance report with workload, member contribution, velocity by team, and workload trend. For person-specific questions, pass userId after using members.",
  inputSchema: z.object({
    filters: analyticsFiltersSchema.optional(),
    userId: z
      .string()
      .optional()
      .describe(
        "Specific user ID for person performance. Use members first when needed.",
      ),
  }),
  execute: async ({ filters, userId }, { experimental_context }) => {
    try {
      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const performance = await getTeamPerformance(
        ctx,
        compactFilters(filters),
      );
      const focusMember = userId
        ? performance.memberContributions.find(
            (member) => member.userId === userId,
          )
        : undefined;

      return {
        success: true,
        kind: "team-performance-report",
        title: focusMember
          ? `${focusMember.username} performance`
          : "Team performance",
        filters: compactFilters(filters),
        userId,
        focusMember,
        performance,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to build team performance report",
      };
    }
  },
});

export const sprintPerformanceReportTool = tool({
  description:
    "Build a sprint performance report across filtered sprints with sprint progress, combined burndown, team allocation, and sprint health.",
  inputSchema: z.object({
    filters: analyticsFiltersSchema.optional(),
  }),
  execute: async ({ filters }, { experimental_context }) => {
    try {
      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const analytics = await getSprintAnalytics(ctx, compactFilters(filters));

      return {
        success: true,
        kind: "sprint-performance-report",
        title: "Sprint performance",
        filters: compactFilters(filters),
        analytics,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to build sprint performance report",
      };
    }
  },
});

export const timelineTrendsReportTool = tool({
  description:
    "Build a timeline trends report for story completion, objective progress, team velocity, active users, stories per day, and cycle time.",
  inputSchema: z.object({
    filters: analyticsFiltersSchema.optional(),
  }),
  execute: async ({ filters }, { experimental_context }) => {
    try {
      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const trends = await getTimelineTrends(ctx, compactFilters(filters));

      return {
        success: true,
        kind: "timeline-trends-report",
        title: "Timeline trends",
        filters: compactFilters(filters),
        trends,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to build timeline trends report",
      };
    }
  },
});
