import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { requireToolConfirmation } from "./tool-helpers";

type MayaWorkPlan = {
  run: {
    id: string;
    status: string;
    summary: string;
  };
  actions: Array<{
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
  }>;
};

export const mayaWorkPlanTool = tool({
  description:
    "Ask Maya to plan assignment and calendar time for a story using backend workload and calendar data. Use when the user wants Maya to schedule work, assign a task to the right person, or find time for a story. Requires admin access and explicit confirmation before applying changes.",
  inputSchema: z.object({
    storyId: z.string().describe("Story ID to plan and schedule."),
    confirmed: z
      .boolean()
      .optional()
      .describe(
        "Must be true after the user explicitly confirms that Maya should create and optionally apply the work plan.",
      ),
    autoApply: z
      .boolean()
      .optional()
      .describe(
        "Set true only after explicit user confirmation to apply assignment and calendar scheduling changes.",
      ),
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
  execute: async (
    {
      storyId,
      confirmed,
      autoApply,
      durationMinutes,
      windowStart,
      windowEnd,
      candidateUserIds,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation(
          "create a Maya work plan for this story",
        );
      }

      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to plan work with Maya",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug?: string })
        .workspaceSlug;

      if (!workspaceSlug) {
        return { success: false, error: "Workspace context is required" };
      }

      const ctx = { session, workspaceSlug };
      const workspace = await getWorkspace(ctx);

      if (workspace.userRole !== "admin") {
        return {
          success: false,
          error: "Only admins can ask Maya to assign and schedule work.",
        };
      }

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
          autoApply,
          durationMinutes,
          windowStart,
          windowEnd,
          candidateUserIds,
        },
        ctx,
        { timeout: 30_000 },
      );

      if (response.error?.message) {
        return {
          success: false,
          error: response.error.message,
        };
      }

      return {
        success: true,
        kind: "maya-work-plan",
        plan: response.data,
        message:
          autoApply === true
            ? "Maya created and applied the work plan."
            : "Maya created a work plan. Ask the user to confirm before applying it.",
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
