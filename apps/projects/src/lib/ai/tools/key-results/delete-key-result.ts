import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { deleteKeyResult } from "@/modules/objectives/actions/delete-key-result";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const deleteKeyResultTool = tool({
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
    const workspace = await getWorkspace(session);
    const userRole = workspace.userRole;

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
