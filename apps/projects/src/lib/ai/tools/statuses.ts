import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createStateAction } from "@/lib/actions/states/create";
import { updateStateAction } from "@/lib/actions/states/update";
import { deleteStateAction } from "@/lib/actions/states/delete";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getTeamStatuses } from "@/lib/queries/states/get-team-states";
import { getTeams } from "@/modules/teams/queries/get-teams";
import type { NewState } from "@/lib/actions/states/create";
import type { UpdateState } from "@/lib/actions/states/update";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const statusesTool = tool({
  description:
    "View, create, update, and manage workflow statuses for teams. Use this tool to answer questions about statuses, count statuses, list team statuses, create new workflow states, or modify existing statuses. For team-specific operations, get the team ID from the teams tool first.",
  inputSchema: z.object({
    action: z
      .enum([
        "list-all-statuses",
        "list-team-statuses",
        "get-status-details",
        "create-status",
        "update-status",
        "delete-status",
        "set-default-status",
      ])
      .describe(
        "Action to perform: list-all-statuses (show all statuses across teams), list-team-statuses (show statuses for a specific team, use this to count team statuses), get-status-details (get info about a specific status), create-status (make new status), update-status (modify existing), delete-status (remove), set-default-status (make default)",
      ),
    // For team-specific operations
    teamId: z
      .string()
      .optional()
      .describe(
        "Team ID for team-specific operations like listing team statuses or creating statuses for a team. Get this from the teams tool first.",
      ),
    // For status operations
    statusId: z
      .string()
      .optional()
      .describe("Status ID for get/update/delete operations"),
    statusName: z.string().optional().describe("Status name for lookups"),
    // For creating/updating statuses
    name: z
      .string()
      .optional()
      .describe("Status name (keep it short and clear)"),
    color: z
      .string()
      .optional()
      .describe(
        "Status color (hex format) - suggest colors that work in both dark and light mode",
      ),
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
      .describe("Status category"),
    isDefault: z
      .boolean()
      .optional()
      .describe("Whether this should be the default status"),
  }),
  execute: async ({
    action,
    teamId,
    statusId,
    statusName,
    name,
    color,
    category,
    isDefault,
  }) => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to access statuses",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      const getTeamName = async (teamId: string): Promise<string | null> => {
        const teams = await getTeams(session);
        const team = teams.find((t) => t.id === teamId);
        return team ? team.name : null;
      };

      const findStatusByName = async (statusName: string, teamId?: string) => {
        const statuses = teamId
          ? await getTeamStatuses(teamId, session)
          : await getStatuses(session);

        const status = statuses.find(
          (s) => s.name.toLowerCase() === statusName.toLowerCase(),
        );
        if (!status) {
          const availableStatuses = statuses.map((s) => s.name).join(", ");
          throw new Error(
            `Status "${statusName}" not found. Available statuses: ${availableStatuses}`,
          );
        }
        return status;
      };

      switch (action) {
        case "list-all-statuses": {
          const statuses = await getStatuses(session);
          return {
            success: true,
            data: statuses.map((status) => ({
              id: status.id,
              name: status.name,
              color: status.color,
              category: status.category,
              isDefault: status.isDefault,
              orderIndex: status.orderIndex,
              teamId: status.teamId,
              createdAt: status.createdAt,
              updatedAt: status.updatedAt,
            })),
            message: `Found ${statuses.length} statuses across all teams`,
          };
        }

        case "list-team-statuses": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for list-team-statuses",
            };
          }

          const teamName = await getTeamName(teamId);
          if (!teamName) {
            return {
              success: false,
              error: `Team with ID "${teamId}" not found`,
            };
          }

          const statuses = await getTeamStatuses(teamId, session);

          return {
            success: true,
            data: statuses.map((status) => ({
              id: status.id,
              name: status.name,
              color: status.color,
              category: status.category,
              isDefault: status.isDefault,
              orderIndex: status.orderIndex,
              createdAt: status.createdAt,
              updatedAt: status.updatedAt,
            })),
            message: `Found ${statuses.length} statuses for team "${teamName}"`,
          };
        }

        case "get-status-details": {
          let status;

          if (statusId) {
            const allStatuses = await getStatuses(session);
            status = allStatuses.find((s) => s.id === statusId);
            if (!status) {
              return {
                success: false,
                error: `Status with ID "${statusId}" not found`,
              };
            }
          } else if (statusName) {
            status = await findStatusByName(statusName, teamId);
          } else {
            return {
              success: false,
              error: "Either statusId or statusName is required",
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
              teamId: status.teamId,
              workspaceId: status.workspaceId,
              createdAt: status.createdAt,
              updatedAt: status.updatedAt,
            },
            message: `Status details for "${status.name}"`,
          };
        }

        case "create-status": {
          if (userRole === "guest") {
            return {
              success: false,
              error:
                "Guests cannot create statuses. Only team members and admins can create statuses.",
            };
          }

          if (!teamId || !name || !category || !color) {
            return {
              success: false,
              error:
                "Team ID, status name, category, and color are required for creating a status",
            };
          }

          const teamName = await getTeamName(teamId);
          if (!teamName) {
            return {
              success: false,
              error: `Team with ID "${teamId}" not found`,
            };
          }

          const newStatusData: NewState = {
            name,
            category,
            teamId,
            color,
          };

          const result = await createStateAction(newStatusData);

          if (result.error) {
            return {
              success: false,
              error: result.error.message,
            };
          }

          return {
            success: true,
            data: result.data,
            message: `Created status "${name}" in team "${teamName}"`,
          };
        }

        case "update-status": {
          if (userRole === "guest") {
            return {
              success: false,
              error:
                "Guests cannot update statuses. Only team members and admins can update statuses.",
            };
          }

          let targetStatusId = statusId;

          if (!targetStatusId && statusName) {
            const status = await findStatusByName(statusName, teamId);
            targetStatusId = status.id;
          }

          if (!targetStatusId) {
            return {
              success: false,
              error: "Either statusId or statusName is required for updating",
            };
          }

          const updateData: UpdateState = {};
          if (name) updateData.name = name;
          if (isDefault !== undefined) updateData.isDefault = isDefault;

          const result = await updateStateAction(targetStatusId, updateData);

          if (result.error) {
            return {
              success: false,
              error: result.error.message,
            };
          }

          return {
            success: true,
            data: result.data,
            message: `Updated status "${statusName || targetStatusId}"`,
          };
        }

        case "delete-status": {
          if (userRole !== "admin") {
            return {
              success: false,
              error: "Only administrators can delete statuses.",
            };
          }

          let targetStatusId = statusId;

          if (!targetStatusId && statusName) {
            const status = await findStatusByName(statusName, teamId);
            targetStatusId = status.id;
          }

          if (!targetStatusId) {
            return {
              success: false,
              error: "Either statusId or statusName is required for deletion",
            };
          }

          const result = await deleteStateAction(targetStatusId);

          if (result.error) {
            return {
              success: false,
              error: result.error.message,
            };
          }

          return {
            success: true,
            message: `Deleted status "${statusName || targetStatusId}"`,
          };
        }

        case "set-default-status": {
          if (userRole === "guest") {
            return {
              success: false,
              error:
                "Guests cannot set default statuses. Only team members and admins can modify default statuses.",
            };
          }

          let targetStatusId = statusId;

          if (!targetStatusId && statusName) {
            const status = await findStatusByName(statusName, teamId);
            targetStatusId = status.id;
          }

          if (!targetStatusId) {
            return {
              success: false,
              error:
                "Either statusId or statusName is required for setting default",
            };
          }

          const result = await updateStateAction(targetStatusId, {
            isDefault: true,
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message,
            };
          }

          return {
            success: true,
            data: result.data,
            message: `Set "${statusName || targetStatusId}" as the default status`,
          };
        }
      }
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "An unknown error occurred",
      };
    }
  },
});
