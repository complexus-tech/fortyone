import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { addAssociationAction } from "@/modules/story/actions/add-association";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const addStoryAssociation = tool({
  description:
    "Associate two stories together (e.g., related, blocking, duplicate).",
  inputSchema: z.object({
    fromStoryId: z.string().describe("The ID of the source story"),
    toStoryId: z
      .string()
      .describe("The ID of the target story to associate with"),
    type: z
      .enum(["related", "blocking", "duplicate"])
      .describe("The type of association"),
  }),
  execute: async (({ fromStoryId, toStoryId, type }), { experimental_context }) => {
    try {
      const session = await auth();
      if (!session) return { success: false, error: "Authentication required" };
      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };


      const workspace = await getWorkspace(session);
      if (workspace.userRole === "guest")
        return { success: false, error: "Unauthorized" };

      const result = await addAssociationAction(fromStoryId, {
        toStoryId,
        type,
      });
      if (result.error) return { success: false, error: result.error.message };

      return { success: true, message: "Association added successfully" };
    } catch (error) {
      return { success: false, error: "Failed to add association" };
    }
  },
});
