import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { removeAssociationAction } from "@/modules/story/actions/remove-association";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const removeStoryAssociation = tool({
  description: "Remove an existing association between two stories.",
  inputSchema: z.object({
    associationId: z.string().describe("The ID of the association to remove"),
  }),
  execute: async (({ associationId }), { experimental_context }) => {
    try {
      const session = await auth();
      if (!session) return { success: false, error: "Authentication required" };
      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };


      const workspace = await getWorkspace(session);
      if (workspace.userRole === "guest")
        return { success: false, error: "Unauthorized" };

      const result = await removeAssociationAction(associationId);
      if (result.error) return { success: false, error: result.error.message };

      return { success: true, message: "Association removed successfully" };
    } catch (error) {
      return { success: false, error: "Failed to remove association" };
    }
  },
});
