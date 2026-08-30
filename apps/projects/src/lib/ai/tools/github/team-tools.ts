import { tool } from "ai";
import { z } from "zod";
import { updateGitHubTeamSettingsAction } from "@/lib/actions/github/update-team-settings";
import { getGitHubTeamSettings } from "@/lib/queries/github/get-team-settings";
import { teamRuleSchema } from "./contracts";
import {
  apiErrorMessage,
  getAuthenticatedGitHubContext,
  getWorkspaceTeams,
  requireConfirmation,
  resolveTeam,
  toUnexpectedToolError,
} from "./helpers";

const teamSelectionSchema = z.object({
  teamId: z.string().optional().describe("FortyOne team ID."),
  teamName: z.string().optional().describe("FortyOne team name."),
});

export const getGitHubTeamSettingsTool = tool({
  description:
    "Get GitHub automation settings and workflow rules for a FortyOne team.",
  inputSchema: teamSelectionSchema,
  execute: async ({ teamId, teamName }, options) => {
    try {
      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const teams = await getWorkspaceTeams(ctx);
      const team = resolveTeam(teams, teamId, teamName);

      if (!team) {
        return {
          success: false,
          error:
            "Team not found. Ask the user for the exact team before reading GitHub settings.",
        };
      }

      const settings = await getGitHubTeamSettings(team.id, ctx);

      return {
        success: true,
        kind: "github-team-automation-report",
        title: `${team.name} GitHub automation`,
        team: {
          id: team.id,
          name: team.name,
          code: team.code,
          color: team.color,
        },
        settings,
        rules: settings.rules.map((rule) => ({
          id: rule.id,
          eventKey: rule.eventKey,
          targetStatusId: rule.targetStatusId,
          baseBranchPattern: rule.baseBranchPattern,
          isActive: rule.isActive,
        })),
        message: `${team.name} has ${settings.rules.length} GitHub automation rule${settings.rules.length === 1 ? "" : "s"}.`,
      };
    } catch (error) {
      return toUnexpectedToolError(error, "Failed to get GitHub team settings");
    }
  },
});

export const updateGitHubTeamSettingsTool = tool({
  description:
    "Replace GitHub automation rules for a FortyOne team. Read existing settings first and send the complete desired rules array.",
  inputSchema: teamSelectionSchema.extend({
    rules: z.array(teamRuleSchema),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms changes."),
  }),
  execute: async ({ teamId, teamName, rules, confirmed }, options) => {
    try {
      if (!confirmed) {
        return requireConfirmation(
          "update this team's GitHub automation rules",
        );
      }

      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const teams = await getWorkspaceTeams(ctx);
      const team = resolveTeam(teams, teamId, teamName);

      if (!team) {
        return {
          success: false,
          error:
            "Team not found. Ask the user for the exact team before updating GitHub settings.",
        };
      }

      const result = await updateGitHubTeamSettingsAction(
        team.id,
        { rules },
        ctx.workspaceSlug,
      );

      if (result.error || !result.data) {
        return {
          success: false,
          error: apiErrorMessage(
            result,
            "Failed to update GitHub team settings",
          ),
        };
      }

      return {
        success: true,
        settings: result.data,
        message: `${team.name} GitHub automation rules updated.`,
      };
    } catch (error) {
      return toUnexpectedToolError(
        error,
        "Failed to update GitHub team settings",
      );
    }
  },
});
