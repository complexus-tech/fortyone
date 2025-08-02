import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getLabels } from "@/lib/queries/labels/get-labels";
import { createLabelAction } from "@/lib/actions/labels/create-label";
import { editLabelAction } from "@/lib/actions/labels/edit-label";
import { deleteLabelAction } from "@/lib/actions/labels/delete-label";
import { getCurrentWorkspace } from "@/lib/hooks/workspaces";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";

export const labelsTool = tool({
  description: "Manage workspace/team labels: list, create, edit, delete.",
  parameters: z.object({
    action: z.enum([
      "list-labels",
      "create-label",
      "edit-label",
      "delete-label",
    ]),
    labelId: z.string().optional(),
    name: z.string().optional(),
    color: z.string().optional(),
    teamId: z.string().optional(),
  }),
  execute: async ({ action, labelId, name, color, teamId }) => {
    try {
      const session = await auth();
      if (!session) return { success: false, error: "Authentication required" };

      // Use workspace userRole only
      const workspaces = await getWorkspaces(session.token);
      const workspace = getCurrentWorkspace(workspaces);
      const userRole = workspace?.userRole;
      const isGuest = userRole === "guest";

      if (isGuest && action !== "list-labels") {
        return { success: false, error: "Guests cannot manage labels" };
      }

      switch (action) {
        case "list-labels":
          return {
            success: true,
            labels: await getLabels(session, teamId ? { teamId } : {}),
          };

        case "create-label":
          if (!name || !color)
            return { success: false, error: "Name and color are required" };
          {
            const res = await createLabelAction({ name, color, teamId });
            if (res.error) return { success: false, error: res.error.message };
            return { success: true, label: res.data };
          }

        case "edit-label":
          if (!labelId || (!name && !color))
            return {
              success: false,
              error: "Label ID and at least one of name or color are required",
            };
          {
            // Only pass defined fields, and ensure both are string if present
            const updates: { name?: string; color?: string } = {};
            if (typeof name === "string") updates.name = name;
            if (typeof color === "string") updates.color = color;
            const res = await editLabelAction(
              labelId,
              updates as { name: string; color: string },
            );
            if (res.error) return { success: false, error: res.error.message };
            return { success: true, label: res.data };
          }

        case "delete-label":
          if (!labelId)
            return { success: false, error: "Label ID is required" };
          {
            const res = await deleteLabelAction(labelId);
            if (res.error) return { success: false, error: res.error.message };
            return { success: true, message: "Label deleted" };
          }

        default:
          return { success: false, error: "Invalid action" };
      }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "An error occurred",
      };
    }
  },
});
