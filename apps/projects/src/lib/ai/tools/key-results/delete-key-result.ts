import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { deleteKeyResult } from "@/modules/objectives/actions/delete-key-result";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const deleteKeyResultTool = tool({
  description:
    "Delete key results from objectives. This action cannot be undone.",
  inputSchema: z.object({
    keyResultId: z.string().describe("Key result ID to delete"),
  }),

  execute: async ({ keyResultId }, { experimental_context }) => {
    const session = await auth();

    if (!session) {

      return {
        success: false,
        error: "Authentication required",
      };
    }

    const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

    const ctx = { session, workspaceSlug };

    // Get user's workspace and role for permissions
    const workspace = await getWorkspace(ctx);
    const userRole = workspace.userRole;

    if (userRole === "guest") {
      return {
        success: false,
        error: "Guests cannot delete key results",
      };
    }

    const result = await deleteKeyResult(keyResultId, workspaceSlug);

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
