import { z } from "zod";
import { tool } from "ai";
import { headers } from "next/headers";
import { auth } from "@/auth";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";
import { getTeamObjectives } from "@/modules/objectives/queries/get-team-objectives";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { getObjectiveAnalytics } from "@/modules/objectives/queries/get-objective-analytics";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { createObjective } from "@/modules/objectives/actions/create-objective";
import { updateObjective } from "@/modules/objectives/actions/update-objective";
import { deleteObjective } from "@/modules/objectives/actions/delete-objective";

export const objectivesTool = tool({
  description:
    "Comprehensive objective management: create, update, delete objectives, track progress, get analytics. Supports team name resolution and role-based permissions.",
  parameters: z.object({
    action: z
      .enum([
        "list-objectives",
        "list-team-objectives",
        "get-objective-details",
        "create-objective",
        "update-objective",
        "delete-objective",
        "get-objective-analytics",
        "get-objectives-overview",
      ])
      .describe("The objective action to perform"),

    teamId: z
      .string()
      .optional()
      .describe("Team ID for filtering objectives by team"),

    objectiveId: z
      .string()
      .optional()
      .describe("Objective ID for single objective operations"),

    objectiveData: z
      .object({
        name: z.string().describe("Objective name (required)"),
        description: z.string().optional().describe("Objective description"),
        teamId: z
          .string()
          .optional()
          .describe(
            "Team ID (optional - will auto-select if user has only one team)",
          ),
        leadUser: z.string().optional().describe("Lead user ID (UUID)"),
        startDate: z.string().describe("Start date (ISO string, required)"),
        endDate: z.string().describe("End date (ISO string, required)"),
        priority: z
          .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
          .optional()
          .describe("Objective priority"),
        statusId: z.string().optional().describe("Status ID"),
        keyResults: z
          .array(
            z.object({
              name: z.string().describe("Key result name"),
              measurementType: z
                .enum(["number", "percentage", "boolean"])
                .describe("How the key result is measured"),
              startValue: z.number().describe("Starting value"),
              targetValue: z.number().describe("Target value"),
            }),
          )
          .optional()
          .describe("Key results to create with the objective"),
      })
      .optional()
      .describe("Objective data for creation"),

    updateData: z
      .object({
        name: z.string().optional().describe("Updated objective name"),
        description: z.string().optional().describe("Updated description"),
        health: z
          .enum(["On Track", "At Risk", "Off Track"])
          .optional()
          .describe("Updated health status"),
      })
      .optional()
      .describe("Objective update data"),

    includeAnalytics: z
      .boolean()
      .optional()
      .describe("Include analytics data in responses"),
  }),

  execute: async ({
    action,
    teamId,
    objectiveId,
    objectiveData,
    updateData,
    includeAnalytics = false,
  }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access objectives",
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
        case "list-objectives": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot view objectives",
            };
          }

          const objectives = await getObjectives(session);

          return {
            success: true,
            objectives: objectives.map((objective) => ({
              id: objective.id,
              name: objective.name,
              description: objective.description,
              teamId: objective.teamId,
              leadUser: objective.leadUser,
              startDate: objective.startDate,
              endDate: objective.endDate,
              statusId: objective.statusId,
              priority: objective.priority,
              health: objective.health,
              createdAt: objective.createdAt,
              updatedAt: objective.updatedAt,
            })),
            count: objectives.length,
            message: `Found ${objectives.length} objectives.`,
          };
        }

        case "list-team-objectives": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for list-team-objectives action",
            };
          }

          if (userRole !== "admin") {
            const userTeams = await getTeams(session);
            const hasAccess = userTeams.some((team) => team.id === teamId);

            if (!hasAccess) {
              return {
                success: false,
                error: "You can only view objectives for teams you belong to",
              };
            }
          }

          const objectives = await getTeamObjectives(teamId, session);

          return {
            success: true,
            objectives: objectives.map((objective) => ({
              id: objective.id,
              name: objective.name,
              description: objective.description,
              teamId: objective.teamId,
              leadUser: objective.leadUser,
              startDate: objective.startDate,
              endDate: objective.endDate,
              statusId: objective.statusId,
              priority: objective.priority,
              health: objective.health,
              createdAt: objective.createdAt,
              updatedAt: objective.updatedAt,
            })),
            count: objectives.length,
            message: `Found ${objectives.length} objectives for team.`,
          };
        }

        case "get-objective-details": {
          if (!objectiveId) {
            return {
              success: false,
              error:
                "Objective ID is required for get-objective-details action",
            };
          }

          const objective = await getObjective(objectiveId, session);

          if (!objective) {
            return {
              success: false,
              error: "Objective not found",
            };
          }

          if (userRole !== "admin") {
            const userTeams = await getTeams(session);
            const hasAccess = userTeams.some(
              (team) => team.id === objective.teamId,
            );

            if (!hasAccess) {
              return {
                success: false,
                error: "You can only view objectives for teams you belong to",
              };
            }
          }

          let analytics;

          if (includeAnalytics) {
            analytics = await getObjectiveAnalytics(objective.id, session);
          }

          return {
            success: true,
            objective: {
              id: objective.id,
              name: objective.name,
              description: objective.description,
              teamId: objective.teamId,
              leadUser: objective.leadUser,
              startDate: objective.startDate,
              endDate: objective.endDate,
              statusId: objective.statusId,
              priority: objective.priority,
              health: objective.health,
              createdAt: objective.createdAt,
              updatedAt: objective.updatedAt,
              ...(analytics ? { analytics } : {}),
            },
            message: `Retrieved objective "${objective.name}".`,
          };
        }

        case "create-objective": {
          if (!objectiveData) {
            return {
              success: false,
              error: "Objective data is required for create-objective action",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot create objectives",
            };
          }

          if (!objectiveData.name) {
            return {
              success: false,
              error: "Objective name is required",
            };
          }

          if (!objectiveData.startDate || !objectiveData.endDate) {
            return {
              success: false,
              error: "Start date and end date are required",
            };
          }

          // Handle team selection
          let finalTeamId = objectiveData.teamId;

          if (!finalTeamId) {
            const userTeams = await getTeams(session);
            if (userTeams.length === 1) {
              finalTeamId = userTeams[0].id;
            } else {
              return {
                success: false,
                error:
                  "Team must be specified when user belongs to multiple teams",
              };
            }
          }

          if (userRole !== "admin") {
            const userTeams = await getTeams(session);
            const hasAccess = userTeams.some((team) => team.id === finalTeamId);

            if (!hasAccess) {
              return {
                success: false,
                error: "You can only create objectives for teams you belong to",
              };
            }
          }

          // Get default status
          const statuses = await getObjectiveStatuses(session);
          const defaultStatus = statuses.find((status) => status.isDefault);
          const statusId =
            objectiveData.statusId || defaultStatus?.id || statuses[0]?.id;

          if (!statusId) {
            return {
              success: false,
              error: "No valid objective status found",
            };
          }

          const result = await createObjective({
            name: objectiveData.name,
            description: objectiveData.description,
            teamId: finalTeamId,
            leadUser: null,
            startDate: objectiveData.startDate,
            endDate: objectiveData.endDate,
            statusId,
            priority: objectiveData.priority,
            keyResults: objectiveData.keyResults?.map((kr) => ({
              name: kr.name,
              measurementType: kr.measurementType,
              startValue: kr.startValue,
              targetValue: kr.targetValue,
              currentValue: kr.startValue,
            })),
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to create objective",
            };
          }

          return {
            success: true,
            objective: result.data,
            message: `Successfully created objective "${objectiveData.name}".`,
          };
        }

        case "update-objective": {
          if (!objectiveId) {
            return {
              success: false,
              error: "Objective ID is required for update-objective action",
            };
          }

          if (!updateData) {
            return {
              success: false,
              error: "Update data is required for update-objective action",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot update objectives",
            };
          }

          const objective = await getObjective(objectiveId, session);
          if (!objective) {
            return {
              success: false,
              error: "Objective not found",
            };
          }

          if (userRole !== "admin") {
            const userTeams = await getTeams(session);
            const hasAccess = userTeams.some(
              (team) => team.id === objective.teamId,
            );

            if (!hasAccess) {
              return {
                success: false,
                error: "You can only update objectives for teams you belong to",
              };
            }
          }

          const result = await updateObjective(objectiveId, {
            name: updateData.name,
            description: updateData.description,
            health: updateData.health,
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to update objective",
            };
          }

          return {
            success: true,
            message: "Successfully updated objective.",
          };
        }

        case "delete-objective": {
          if (!objectiveId) {
            return {
              success: false,
              error: "Objective ID is required for delete-objective action",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot delete objectives",
            };
          }

          const objective = await getObjective(objectiveId, session);
          if (!objective) {
            return {
              success: false,
              error: "Objective not found",
            };
          }

          if (
            userRole !== "admin" &&
            objective.createdBy !== session.user?.id
          ) {
            return {
              success: false,
              error: "You can only delete objectives you created",
            };
          }

          const result = await deleteObjective(objectiveId);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to delete objective",
            };
          }

          return {
            success: true,
            message: "Successfully deleted objective.",
          };
        }

        case "get-objective-analytics": {
          if (!objectiveId) {
            return {
              success: false,
              error:
                "Objective ID is required for get-objective-analytics action",
            };
          }

          const objective = await getObjective(objectiveId, session);
          if (!objective) {
            return {
              success: false,
              error: "Objective not found",
            };
          }

          if (userRole !== "admin") {
            const userTeams = await getTeams(session);
            const hasAccess = userTeams.some(
              (team) => team.id === objective.teamId,
            );

            if (!hasAccess) {
              return {
                success: false,
                error: "You can only view analytics for teams you belong to",
              };
            }
          }

          const analytics = await getObjectiveAnalytics(objectiveId, session);

          return {
            success: true,
            analytics,
            message: `Retrieved analytics for objective "${objective.name}".`,
          };
        }

        case "get-objectives-overview": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot view objectives overview",
            };
          }

          const objectives = await getObjectives(session);

          const total = objectives.length;
          const byHealth = {
            onTrack: objectives.filter((obj) => obj.health === "On Track")
              .length,
            atRisk: objectives.filter((obj) => obj.health === "At Risk").length,
            offTrack: objectives.filter((obj) => obj.health === "Off Track")
              .length,
            notSet: objectives.filter((obj) => !obj.health).length,
          };

          const byPriority = {
            urgent: objectives.filter((obj) => obj.priority === "Urgent")
              .length,
            high: objectives.filter((obj) => obj.priority === "High").length,
            medium: objectives.filter((obj) => obj.priority === "Medium")
              .length,
            low: objectives.filter((obj) => obj.priority === "Low").length,
            noPriority: objectives.filter(
              (obj) => obj.priority === "No Priority" || !obj.priority,
            ).length,
          };

          return {
            success: true,
            overview: {
              total,
              health: byHealth,
              priority: byPriority,
              recentObjectives: objectives
                .sort(
                  (a, b) =>
                    new Date(b.createdAt).getTime() -
                    new Date(a.createdAt).getTime(),
                )
                .slice(0, 5)
                .map((obj) => ({
                  id: obj.id,
                  name: obj.name,
                  health: obj.health,
                  priority: obj.priority,
                  createdAt: obj.createdAt,
                })),
            },
            message: `Overview of ${total} objectives.`,
          };
        }
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
