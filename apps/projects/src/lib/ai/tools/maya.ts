import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { getApiError } from "@/utils";
import { resolveMayaWorkPlanResponse } from "./maya-work-plan-response";

type MayaWorkPlan = {
  run: {
    id: string;
    status: string;
    summary: string;
  };
  actions: {
    id: string;
    type: "assign_story" | "schedule_work_block" | "flag_schedule_risk";
    status: "proposed" | "applied" | "failed";
    reason: string;
    payload: {
      assignStory?: { assigneeId: string };
      scheduleBlock?: {
        userId: string;
        title: string;
        startAt: string;
        endAt: string;
      };
      risk?: {
        code: string;
        message: string;
      };
    };
  }[];
};

type MayaWorkPlanToolResult = {
  error?: string;
  kind?: "maya-work-plan";
  message?: string;
  phase?: "applied" | "preview";
  plan?: MayaWorkPlan;
  success: boolean;
};

const toMayaWorkPlanModelOutput = ({ output }: { output: unknown }) => {
  const result = output as MayaWorkPlanToolResult;
  const plan = result.plan
    ? {
        actions: result.plan.actions.map((action) => ({
          id: action.id,
          payload: {
            ...(action.payload.assignStory
              ? {
                  assignStory: {
                    assigneeId: action.payload.assignStory.assigneeId,
                  },
                }
              : {}),
            ...(action.payload.scheduleBlock
              ? {
                  scheduleBlock: {
                    endAt: action.payload.scheduleBlock.endAt,
                    startAt: action.payload.scheduleBlock.startAt,
                    title: action.payload.scheduleBlock.title,
                    userId: action.payload.scheduleBlock.userId,
                  },
                }
              : {}),
            ...(action.payload.risk ? { risk: action.payload.risk } : {}),
          },
          reason: action.reason,
          status: action.status,
          type: action.type,
        })),
        run: {
          id: result.plan.run.id,
          status: result.plan.run.status,
          summary: result.plan.run.summary,
        },
      }
    : undefined;

  return {
    type: "json" as const,
    value: {
      success: result.success,
      ...(result.error ? { error: result.error } : {}),
      ...(result.kind ? { kind: result.kind } : {}),
      ...(result.message ? { message: result.message } : {}),
      ...(result.phase ? { phase: result.phase } : {}),
      ...(plan ? { plan } : {}),
    },
  };
};

const getMemberMayaContext = async (experimentalContext: unknown) => {
  const session = await auth();
  if (!session) {
    return {
      error: "Authentication required to plan work with Maya",
    } as const;
  }

  const workspaceSlug = (experimentalContext as { workspaceSlug?: string })
    .workspaceSlug;
  if (!workspaceSlug) {
    return { error: "Workspace context is required" } as const;
  }

  const ctx = { session, workspaceSlug };
  const workspace = await getWorkspace(ctx);
  if (workspace.userRole !== "admin" && workspace.userRole !== "member") {
    return {
      error: "Only workspace admins and members can assign and schedule work.",
    } as const;
  }

  return { ctx } as const;
};

export const mayaWorkPlanTool = tool({
  description:
    "Prepare a non-mutating assignment and calendar work-plan preview using current workload and availability. Use this before applyMayaWorkPlanTool whenever a workspace admin or member wants Maya to assign or schedule a story. Present the preview, then pass its exact run ID to the apply tool; never recompute it.",
  inputSchema: z.object({
    storyId: z.string().describe("Story ID to plan and schedule."),
    durationMinutes: z
      .number()
      .int()
      .positive()
      .optional()
      .describe("Optional work duration in minutes."),
    windowStart: z
      .string()
      .optional()
      .describe("Optional ISO datetime for the planning window start."),
    windowEnd: z
      .string()
      .optional()
      .describe("Optional ISO datetime for the planning window end."),
    candidateUserIds: z
      .array(z.string())
      .optional()
      .describe("Optional list of candidate user IDs Maya may choose from."),
  }),
  toModelOutput: toMayaWorkPlanModelOutput,
  execute: async (
    { storyId, durationMinutes, windowStart, windowEnd, candidateUserIds },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const context = await getMemberMayaContext(experimentalContext);
      if ("error" in context) return { success: false, error: context.error };

      const response = await post<
        {
          storyId: string;
          autoApply?: boolean;
          durationMinutes?: number;
          windowStart?: string;
          windowEnd?: string;
          candidateUserIds?: string[];
        },
        ApiResponse<MayaWorkPlan>
      >(
        "maya/work-plans",
        {
          storyId,
          autoApply: false,
          durationMinutes,
          windowStart,
          windowEnd,
          candidateUserIds,
        },
        context.ctx,
        { timeout: 30_000 },
      );

      const resolved = resolveMayaWorkPlanResponse(
        response,
        "Maya returned no work-plan preview. Please try again.",
      );
      if ("error" in resolved) return { success: false, error: resolved.error };

      return {
        success: true,
        kind: "maya-work-plan",
        phase: "preview",
        plan: resolved.data,
        message: "Maya prepared this work plan for approval.",
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to create Maya work plan",
      };
    }
  },
});

export const applyMayaWorkPlanTool = tool({
  description:
    "Apply the exact persisted Maya work-plan preview identified by runId. Use only with the run ID returned by mayaWorkPlanTool. This mutation pauses for native interface approval and never recalculates the plan.",
  inputSchema: z.object({
    runId: z.string().uuid().describe("Exact run ID from mayaWorkPlanTool."),
  }),
  toModelOutput: toMayaWorkPlanModelOutput,
  execute: async ({ runId }, { experimental_context: experimentalContext }) => {
    try {
      const context = await getMemberMayaContext(experimentalContext);
      if ("error" in context) return { success: false, error: context.error };

      const response = await post<
        Record<string, never>,
        ApiResponse<MayaWorkPlan>
      >(`maya/work-plans/${runId}/apply`, {}, context.ctx, { timeout: 30_000 });
      const resolved = resolveMayaWorkPlanResponse(
        response,
        "Maya returned no applied work plan. Please try again.",
      );
      if ("error" in resolved) return { success: false, error: resolved.error };

      const actions = resolved.data.actions;
      const failedCount = actions.filter(
        (action) => action.status === "failed",
      ).length;
      return {
        success: failedCount === 0,
        kind: "maya-work-plan",
        phase: "applied",
        plan: resolved.data,
        message:
          failedCount === 0
            ? "Maya applied the approved work plan."
            : `Maya could not apply ${failedCount} work-plan ${failedCount === 1 ? "action" : "actions"}.`,
      };
    } catch (error) {
      const result = getApiError(error);
      return {
        success: false,
        error: result.error?.message || "Failed to apply Maya work plan",
      };
    }
  },
});
