import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getPublicTeams } from "@/modules/teams/queries/get-public-teams";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { createTeam } from "@/modules/teams/actions";
import { updateTeamAction } from "@/modules/teams/actions/update-team";
import { addTeamMemberAction } from "@/modules/teams/actions/add-team-member";
import { removeTeamMemberAction } from "@/modules/teams/actions/remove-team-member";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { getCurrentWorkspace } from "@/lib/hooks/workspaces";
import { deleteTeamAction } from "@/modules/teams/actions/delete-team";

export const teamsTool = tool({
  description:
    "Manage teams, view members, and handle team operations based on user permissions",
  parameters: z.object({
    action: z
      .enum([
        "list-teams",
        "list-public-teams",
        "get-team-details",
        "list-members",
        "create-team",
        "update-team",
        "join-team",
        "delete-team",
        "leave-team",
      ])
      .describe("The action to perform on teams"),
    teamId: z
      .string()
      .optional()
      .describe("Team ID for specific team operations"),
    teamData: z
      .object({
        name: z.string().describe("Team name"),
        code: z.string().describe("Team code (unique identifier)"),
        color: z.string().describe("Team color (hex code)"),
        isPrivate: z.boolean().optional().describe("Whether team is private"),
      })
      .optional()
      .describe("Team data for create/update operations"),
    updateData: z
      .object({
        name: z.string().optional().describe("Updated team name"),
        color: z.string().optional().describe("Updated team color"),
        isPrivate: z.boolean().optional().describe("Updated privacy setting"),
      })
      .optional()
      .describe("Data for updating team"),
  }),
  execute: async ({ action, teamId, teamData, updateData }) => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to access teams",
        };
      }
      const workspaces = await getWorkspaces(session.token);
      const workspace = getCurrentWorkspace(workspaces);
      const userRole = workspace?.userRole;
      const currentUserId = session.user?.id;

      if (!userRole) {
        return {
          success: false,
          error: "Unable to determine user permissions",
        };
      }

      switch (action) {
        case "list-teams": {
          const teams = await getTeams(session);
          return {
            success: true,
            teams: teams.map((team) => ({
              id: team.id,
              name: team.name,
              code: team.code,
              memberCount: team.memberCount,
              isPrivate: team.isPrivate,
              color: team.color,
            })),
            count: teams.length,
            userRole,
            message:
              teams.length > 0
                ? `You are a member of ${teams.length} team${teams.length !== 1 ? "s" : ""}: ${teams.map((t) => t.name).join(", ")}.`
                : "You are not currently a member of any teams.",
          };
        }

        case "list-public-teams": {
          const publicTeams = await getPublicTeams(session);
          const userTeams = await getTeams(session);
          const userTeamIds = userTeams.map((t) => t.id);
          const availableTeams = publicTeams.filter(
            (team) => !userTeamIds.includes(team.id),
          );

          return {
            success: true,
            teams: availableTeams.map((team) => ({
              id: team.id,
              name: team.name,
              code: team.code,
              memberCount: team.memberCount,
              color: team.color,
            })),
            count: availableTeams.length,
            message:
              availableTeams.length > 0
                ? `There are ${availableTeams.length} public team${availableTeams.length !== 1 ? "s" : ""} you can join: ${availableTeams.map((t) => t.name).join(", ")}.`
                : "No public teams available to join.",
          };
        }
        case "get-team-details": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for get-team-details action",
            };
          }
          const teams = await getTeams(session);
          const team = teams.find((t) => t.id === teamId);
          if (!team) {
            return {
              success: false,
              error: "Team not found or you don't have access to this team",
            };
          }
          const members = await getTeamMembers(teamId, session);
          return {
            success: true,
            team: {
              id: team.id,
              name: team.name,
              code: team.code,
              color: team.color,
              isPrivate: team.isPrivate,
              memberCount: team.memberCount,
              createdAt: team.createdAt,
              members: members.map((member) => ({
                id: member.id,
                name: member.fullName,
                username: member.username,
                role: member.role,
                avatarUrl: member.avatarUrl,
              })),
            },
            message: `Here are the details for ${team.name}. It has ${members.length} member${members.length !== 1 ? "s" : ""}.`,
          };
        }
        case "list-members": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for list-members action",
            };
          }

          const teams = await getTeams(session);
          const team = teams.find((t) => t.id === teamId);

          if (!team) {
            return {
              success: false,
              error: "Team not found or you don't have access to this team",
            };
          }

          const members = await getTeamMembers(teamId, session);

          return {
            success: true,
            teamName: team.name,
            members: members.map((member) => ({
              id: member.id,
              name: member.fullName,
              username: member.username,
              role: member.role,
              avatarUrl: member.avatarUrl,
            })),
            count: members.length,
            message: `${team.name} has ${members.length} member${members.length !== 1 ? "s" : ""}: ${members.map((m) => m.fullName).join(", ")}.`,
          };
        }
        case "create-team": {
          if (userRole !== "admin") {
            return {
              success: false,
              error: "Only admins can create teams.",
            };
          }

          if (!teamData) {
            return {
              success: false,
              error: "Team data is required for create-team action",
            };
          }
          const result = await createTeam({
            ...teamData,
            isPrivate: teamData.isPrivate ?? false,
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to create team",
            };
          }

          return {
            success: true,
            team: {
              id: result.data?.id,
              name: result.data?.name,
              code: result.data?.code,
              color: result.data?.color,
              isPrivate: result.data?.isPrivate,
            },
            message: `Successfully created team "${teamData.name}" with code "${teamData.code}". You are now a member of this team.`,
          };
        }
        case "update-team": {
          if (userRole !== "admin") {
            return {
              success: false,
              error: "Only admins can update team settings.",
            };
          }

          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for update-team action",
            };
          }

          if (!updateData) {
            return {
              success: false,
              error: "Update data is required for update-team action",
            };
          }
          const result = await updateTeamAction(teamId, updateData);
          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to update team",
            };
          }
          return {
            success: true,
            team: result.data,
            message: `Successfully updated team settings.`,
          };
        }
        case "join-team": {
          if (userRole === "guest") {
            return {
              success: false,
              error:
                "Guests cannot join teams. Contact an admin to be added to teams.",
            };
          }

          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for join-team action",
            };
          }

          if (!currentUserId) {
            return {
              success: false,
              error: "User ID not found in session",
            };
          }

          // Check if team exists and is public or user has access
          const publicTeams = await getPublicTeams(session);
          const team = publicTeams.find((t) => t.id === teamId);

          if (!team) {
            return {
              success: false,
              error:
                "Team not found or is private. You can only join public teams.",
            };
          }
          const result = await addTeamMemberAction(teamId, currentUserId);
          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to join team",
            };
          }
          return {
            success: true,
            message: `Successfully joined "${team.name}". Welcome to the team!`,
          };
        }
        case "leave-team": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for leave-team action",
            };
          }
          if (!currentUserId) {
            return {
              success: false,
              error: "User ID not found in session",
            };
          }
          const teams = await getTeams(session);
          const team = teams.find((t) => t.id === teamId);
          if (!team) {
            return {
              success: false,
              error: "You are not a member of this team",
            };
          }

          // Prevent leaving if this is the user's only team
          if (teams.length === 1) {
            return {
              success: false,
              error: `You cannot leave "${team.name}" because it's your only team. You must belong to at least one team to use the platform.`,
            };
          }

          const result = await removeTeamMemberAction(teamId, currentUserId);
          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to leave team",
            };
          }
          return {
            success: true,
            message: `Successfully left "${team.name}".`,
          };
        }
        case "delete-team": {
          if (userRole !== "admin") {
            return {
              success: false,
              error: "Only admins can delete teams.",
            };
          }

          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for delete-team action",
            };
          }

          const teams = await getTeams(session);
          const team = teams.find((t) => t.id === teamId);

          if (!team) {
            return {
              success: false,
              error: "Team not found or you don't have access to this team",
            };
          }

          const result = await deleteTeamAction(teamId);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to delete team",
            };
          }

          return {
            success: true,
            message: `Successfully deleted team "${team.name}".`,
          };
        }
      }
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred while performing team operation",
      };
    }
  },
});
