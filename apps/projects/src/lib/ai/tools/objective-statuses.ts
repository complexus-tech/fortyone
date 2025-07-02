import { z } from "zod";
import { tool } from "ai";
import { headers } from "next/headers";
import { auth } from "@/auth";
import { createObjectiveStatusAction } from "@/modules/objectives/actions/statuses/create";
import { updateObjectiveStatusAction } from "@/modules/objectives/actions/statuses/update";
import { deleteObjectiveStatusAction } from "@/modules/objectives/actions/statuses/delete";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import type { NewObjectiveStatus } from "@/modules/objectives/actions/statuses/create";
import type { UpdateObjectiveStatus } from "@/modules/objectives/actions/statuses/update";

export const objectiveStatusesTool = tool({
  description:
    "View, create, update, and manage workflow statuses for objectives. These are workspace-level statuses only (no team-specific statuses). Use this tool to list objective statuses, create new workflow states for objectives, or modify existing objective statuses.",
  parameters: z.object({
    action: z
      .enum([
        "list-objective-statuses",
        "get-objective-status-details",
        "create-objective-status",
        "update-objective-status",
        "delete-objective-status",
        "set-default-objective-status",
      ])
      .describe(
        "Action to perform: list-objective-statuses (show all objective statuses), get-objective-status-details (get info about a specific status), create-objective-status (make new status), update-objective-status (modify existing), delete-objective-status (remove), set-default-objective-status (make default)",
      ),

    // For status operations
    statusId: z
      .string()
      .optional()
      .describe("Objective status ID for get/update/delete operations"),

    // For creating/updating statuses
    name: z.string().optional().describe("Objective status name"),
    color: z
      .string()
      .optional()
      .describe("Objective status color (hex format)"),
    category: z
      .enum([
        "backlog",
        "unstarted",
        "started",
        "paused",
        "completed",
        "cancelled",
      ])
      .optional()
      .describe("Objective status category"),
    isDefault: z
      .boolean()
      .optional()
      .describe("Whether this should be the default objective status"),
  }),
  execute: async ({ action, statusId, name, color, category, isDefault }) => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to access objective statuses",
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

      switch (action) {
        case "list-objective-statuses": {
          const statuses = await getObjectiveStatuses(session);
          return {
            success: true,
            data: statuses.map((status) => ({
              id: status.id,
              name: status.name,
              color: status.color,
              category: status.category,
              isDefault: status.isDefault,
              orderIndex: status.orderIndex,
              workspaceId: status.workspaceId,
              createdAt: status.createdAt,
              updatedAt: status.updatedAt,
            })),
            message: `Found ${statuses.length} objective statuses in workspace`,
          };
        }

        case "get-objective-status-details": {
          if (!statusId) {
            return {
              success: false,
              error:
                "Status ID is required for get-objective-status-details action",
            };
          }

          const allStatuses = await getObjectiveStatuses(session);
          const status = allStatuses.find((s) => s.id === statusId);

          if (!status) {
            return {
              success: false,
              error: `Objective status with ID "${statusId}" not found`,
            };
          }

          return {
            success: true,
            data: {
              id: status.id,
              name: status.name,
              color: status.color,
              category: status.category,
              isDefault: status.isDefault,
              orderIndex: status.orderIndex,
              workspaceId: status.workspaceId,
              createdAt: status.createdAt,
              updatedAt: status.updatedAt,
            },
            message: `Objective status details for "${status.name}"`,
          };
        }

        case "create-objective-status": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot create objective statuses",
            };
          }

          if (!name || !color || !category) {
            return {
              success: false,
              error:
                "Name, color, and category are required for creating objective status",
            };
          }

          const newStatus: NewObjectiveStatus = {
            name,
            color,
            category,
          };

          const result = await createObjectiveStatusAction(newStatus);

          if (result.error) {
            return {
              success: false,
              error:
                result.error.message || "Failed to create objective status",
            };
          }

          return {
            success: true,
            data: result.data,
            message: `Successfully created objective status "${name}"`,
          };
        }

        case "update-objective-status": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot update objective statuses",
            };
          }

          if (!statusId) {
            return {
              success: false,
              error: "Status ID is required for update operation",
            };
          }

          const updateData: UpdateObjectiveStatus = {};
          if (name !== undefined) updateData.name = name;
          if (color !== undefined) updateData.color = color;
          if (category !== undefined) updateData.category = category;
          if (isDefault !== undefined) updateData.isDefault = isDefault;

          if (Object.keys(updateData).length === 0) {
            return {
              success: false,
              error: "At least one field must be provided for update",
            };
          }

          const result = await updateObjectiveStatusAction(
            statusId,
            updateData,
          );

          if (result.error) {
            return {
              success: false,
              error:
                result.error.message || "Failed to update objective status",
            };
          }

          return {
            success: true,
            data: result.data,
            message: `Successfully updated objective status`,
          };
        }

        case "set-default-objective-status": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot set default objective statuses",
            };
          }

          if (!statusId) {
            return {
              success: false,
              error: "Status ID is required to set default status",
            };
          }

          const result = await updateObjectiveStatusAction(statusId, {
            isDefault: true,
          });

          if (result.error) {
            return {
              success: false,
              error:
                result.error.message ||
                "Failed to set default objective status",
            };
          }

          return {
            success: true,
            data: result.data,
            message: `Successfully set default objective status`,
          };
        }

        case "delete-objective-status": {
          if (userRole !== "admin") {
            return {
              success: false,
              error: "Only admins can delete objective statuses",
            };
          }

          if (!statusId) {
            return {
              success: false,
              error: "Status ID is required for delete operation",
            };
          }

          const result = await deleteObjectiveStatusAction(statusId);

          if (result.error) {
            return {
              success: false,
              error:
                result.error.message || "Failed to delete objective status",
            };
          }

          return {
            success: true,
            message: `Successfully deleted objective status`,
          };
        }

        default:
          return {
            success: false,
            error: "Invalid action specified",
          };
      }
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred",
      };
    }
  },
});
