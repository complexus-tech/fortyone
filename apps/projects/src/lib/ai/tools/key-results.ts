import { z } from "zod";
import { tool } from "ai";
import { headers } from "next/headers";
import { auth } from "@/auth";
import { getWorkspaceKeyResults } from "@/modules/key-results/queries/get-workspace-key-results";
import { createKeyResult } from "@/modules/objectives/actions/create-key-result";
import { updateKeyResult } from "@/modules/objectives/actions/update-key-result";
import { deleteKeyResult } from "@/modules/objectives/actions/delete-key-result";
import { getObjective } from "@/modules/objectives/queries/get-objective";

// Tool 1: List Key Results
export const keyResultsListTool = tool({
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
    const headersList = await headers();
    const subdomain = headersList.get("host")?.split(".")[0] || "";
    const workspace = session.workspaces.find(
      (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
    );
    const userRole = workspace?.userRole;

    if (!userRole) {
      return {
        success: false,
        error: "Unable to determine user permissions",
      };
    }

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

export const keyResultsCreateTool = tool({
  description:
    "Create new key results within objectives. Requires objective ID and key result details.",
  parameters: z.object({
    objectiveId: z
      .string()
      .describe("Objective ID where the key result will be created"),
    name: z.string().describe("Key result name"),
    measurementType: z
      .enum(["percentage", "number", "boolean"])
      .describe("How the key result is measured"),
    startValue: z.number().describe("Starting value (0, 1 for boolean)"),
    currentValue: z
      .number()
      .optional()
      .describe(
        "Current progress value (defaults to startValue if not provided)",
      ),
    targetValue: z
      .number()
      .describe("Target value (required for percentage and number types)"),
  }),

  execute: async ({
    objectiveId,
    name,
    measurementType,
    startValue,
    currentValue,
    targetValue,
  }) => {
    const session = await auth();

    if (!session) {
      return {
        success: false,
        error: "Authentication required",
      };
    }

    // Get user's workspace and role for permissions
    const headersList = await headers();
    const subdomain = headersList.get("host")?.split(".")[0] || "";
    const workspace = session.workspaces.find(
      (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
    );
    const userRole = workspace?.userRole;

    if (!userRole) {
      return {
        success: false,
        error: "Unable to determine user permissions",
      };
    }

    if (userRole === "guest") {
      return {
        success: false,
        error: "Guests cannot create key results",
      };
    }

    // Check if user has access to the objective's team
    const objective = await getObjective(objectiveId, session);
    if (!objective) {
      return {
        success: false,
        error: "Objective not found",
      };
    }

    const finalCurrentValue = currentValue ?? startValue;
    const result = await createKeyResult({
      objectiveId,
      name,
      measurementType,
      startValue,
      targetValue,
      currentValue: finalCurrentValue,
    });

    if (result.error) {
      return {
        success: false,
        error: result.error.message || "Failed to create key result",
      };
    }

    return {
      success: true,
      message: `Successfully created key result "${name}".`,
    };
  },
});

// Tool 3: Update Key Result
export const keyResultsUpdateTool = tool({
  description:
    "Update existing key results. Can update name, measurement type, and values.",
  parameters: z.object({
    keyResultId: z.string().describe("Key result ID to update"),
    objectiveId: z.string().describe("Objective ID to update"),
    name: z.string().optional().describe("Updated key result name"),
    startValue: z.number().optional().describe("Updated starting value"),
    currentValue: z
      .number()
      .optional()
      .describe("Updated current progress value"),
    targetValue: z.number().optional().describe("Updated target value"),
  }),

  execute: async ({
    keyResultId,
    name,
    startValue,
    currentValue,
    targetValue,
    objectiveId,
  }) => {
    const session = await auth();

    if (!session) {
      return {
        success: false,
        error: "Authentication required",
      };
    }

    // Get user's workspace and role for permissions
    const headersList = await headers();
    const subdomain = headersList.get("host")?.split(".")[0] || "";
    const workspace = session.workspaces.find(
      (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
    );
    const userRole = workspace?.userRole;

    if (userRole === "guest") {
      return {
        success: false,
        error: "Guests cannot update key results",
      };
    }

    const result = await updateKeyResult(keyResultId, objectiveId, {
      name,
      startValue,
      currentValue,
      targetValue,
    });

    if (result.error) {
      return {
        success: false,
        error: result.error.message || "Failed to update key result",
      };
    }

    return {
      success: true,
      message: "Successfully updated key result.",
    };
  },
});

// Tool 4: Delete Key Result
export const keyResultsDeleteTool = tool({
  description:
    "Delete key results from objectives. This action cannot be undone.",
  parameters: z.object({
    keyResultId: z.string().describe("Key result ID to delete"),
  }),

  execute: async ({ keyResultId }) => {
    const session = await auth();

    if (!session) {
      return {
        success: false,
        error: "Authentication required",
      };
    }

    // Get user's workspace and role for permissions
    const headersList = await headers();
    const subdomain = headersList.get("host")?.split(".")[0] || "";
    const workspace = session.workspaces.find(
      (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
    );
    const userRole = workspace?.userRole;

    if (!userRole) {
      return {
        success: false,
        error: "Unable to determine user permissions",
      };
    }

    if (userRole === "guest") {
      return {
        success: false,
        error: "Guests cannot delete key results",
      };
    }

    const result = await deleteKeyResult(keyResultId);

    if (result.error) {
      return {
        success: false,
        error: result.error.message || "Failed to delete key result",
      };
    }

    return {
      success: true,
      message: "Successfully deleted key result.",
    };
  },
});
