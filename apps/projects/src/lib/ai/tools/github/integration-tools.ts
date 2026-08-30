import { tool } from "ai";
import { z } from "zod";
import { createGitHubInstallSessionAction } from "@/lib/actions/github/create-install-session";
import { createGitHubIssueSyncLinkAction } from "@/lib/actions/github/create-issue-sync-link";
import { deleteGitHubIssueSyncLinkAction } from "@/lib/actions/github/delete-issue-sync-link";
import { resyncGitHubRepositoriesAction } from "@/lib/actions/github/resync-repositories";
import { updateGitHubWorkspaceSettingsAction } from "@/lib/actions/github/update-workspace-settings";
import { getGitHubIntegration } from "@/lib/queries/github/get-integration";
import {
  syncDirectionSchema,
  toRepositorySummary,
  toWorkspaceSettingsUpdate,
  workspaceSettingsUpdateSchema,
} from "./contracts";
import {
  apiErrorMessage,
  getAuthenticatedGitHubContext,
  getWorkspaceTeams,
  requireConfirmation,
  resolveRepository,
  resolveTeam,
  toUnexpectedToolError,
} from "./helpers";

const confirmationSchema = z
  .boolean()
  .optional()
  .describe("Must be true after the user explicitly confirms changes.");

export const getGitHubIntegrationTool = tool({
  description:
    "Get the workspace GitHub integration status, installations, repositories, issue sync links, and workspace GitHub settings.",
  inputSchema: z.object({}),
  execute: async (_input, options) => {
    try {
      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const integration = await getGitHubIntegration(ctx);
      const activeRepositories = integration.repositories.filter(
        (repository) => repository.isActive,
      );
      const activeSyncLinks = integration.issueSyncLinks.filter(
        (link) => link.isActive,
      );

      return {
        success: true,
        kind: "github-integration-report",
        title: "GitHub integration",
        summary: {
          connected: integration.installations.some(
            (installation) => installation.isActive,
          ),
          installations: integration.installations.length,
          repositories: integration.repositories.length,
          activeRepositories: activeRepositories.length,
          issueSyncLinks: integration.issueSyncLinks.length,
          activeIssueSyncLinks: activeSyncLinks.length,
        },
        settings: integration.settings,
        installations: integration.installations.map((installation) => ({
          accountLogin: installation.accountLogin,
          accountType: installation.accountType,
          repositorySelection: installation.repositorySelection,
          isActive: installation.isActive,
        })),
        repositories: integration.repositories.map(toRepositorySummary),
        issueSyncLinks: integration.issueSyncLinks.map((link) => ({
          id: link.id,
          repositoryId: link.repositoryId,
          repositoryName: link.repositoryName,
          teamId: link.teamId,
          teamName: link.teamName,
          teamColor: link.teamColor,
          syncDirection: link.syncDirection,
          isActive: link.isActive,
        })),
        message: integration.installations.length
          ? `GitHub is connected with ${activeRepositories.length} active repositories and ${activeSyncLinks.length} active issue sync links.`
          : "GitHub is not connected for this workspace.",
      };
    } catch (error) {
      return toUnexpectedToolError(error, "Failed to get GitHub integration");
    }
  },
});

export const createGitHubInstallSessionTool = tool({
  description:
    "Create a GitHub App installation session for the current workspace. Use this when GitHub is not connected and the user wants to connect it.",
  inputSchema: z.object({}),
  execute: async (_input, options) => {
    try {
      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const result = await createGitHubInstallSessionAction(ctx.workspaceSlug);

      if (result.error || !result.data?.installUrl) {
        return {
          success: false,
          error: apiErrorMessage(
            result,
            "Failed to create GitHub install session",
          ),
        };
      }

      return {
        success: true,
        installUrl: result.data.installUrl,
        message: "GitHub install session created.",
      };
    } catch (error) {
      return toUnexpectedToolError(
        error,
        "Failed to create GitHub install session",
      );
    }
  },
});

export const resyncGitHubRepositoriesTool = tool({
  description:
    "Resync repositories from the connected GitHub installation for the current workspace.",
  inputSchema: z.object({
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms resyncing."),
  }),
  execute: async ({ confirmed }, options) => {
    try {
      if (!confirmed) {
        return requireConfirmation("resync GitHub repositories");
      }

      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const result = await resyncGitHubRepositoriesAction(ctx.workspaceSlug);

      if (result.error) {
        return {
          success: false,
          error: apiErrorMessage(result, "Failed to resync repositories"),
        };
      }

      return {
        success: true,
        message: "GitHub repositories were resynced.",
      };
    } catch (error) {
      return toUnexpectedToolError(
        error,
        "Failed to resync GitHub repositories",
      );
    }
  },
});

export const createGitHubIssueSyncLinkTool = tool({
  description:
    "Create an issue sync link between a GitHub repository and a FortyOne team. Resolve repository and team by ID or exact name.",
  inputSchema: z.object({
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms linking."),
    repositoryFullName: z
      .string()
      .optional()
      .describe("GitHub repository full name, for example owner/repo."),
    repositoryId: z.string().optional().describe("GitHub repository ID."),
    syncDirection: syncDirectionSchema.default("inbound_only"),
    teamId: z.string().optional().describe("FortyOne team ID."),
    teamName: z.string().optional().describe("FortyOne team name."),
  }),
  execute: async (
    {
      repositoryId,
      repositoryFullName,
      teamId,
      teamName,
      syncDirection,
      confirmed,
    },
    options,
  ) => {
    try {
      if (!confirmed) {
        return requireConfirmation("link this GitHub repository to the team");
      }

      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const [integration, teams] = await Promise.all([
        getGitHubIntegration(ctx),
        getWorkspaceTeams(ctx),
      ]);
      const repository = resolveRepository(
        integration.repositories,
        repositoryId,
        repositoryFullName,
      );

      if (!repository) {
        return {
          success: false,
          error:
            "GitHub repository not found. Ask the user for the exact repository or resync repositories first.",
        };
      }

      const team = resolveTeam(teams, teamId, teamName);
      if (!team) {
        return {
          success: false,
          error:
            "Team not found. Ask the user for the exact team before creating the sync link.",
        };
      }

      const result = await createGitHubIssueSyncLinkAction(
        {
          repositoryId: repository.id,
          teamId: team.id,
          syncDirection,
        },
        ctx.workspaceSlug,
      );

      if (result.error || !result.data) {
        return {
          success: false,
          error: apiErrorMessage(result, "Failed to create issue sync link"),
        };
      }

      return {
        success: true,
        issueSyncLink: result.data,
        message: `${repository.fullName} is now linked to ${team.name} for GitHub issue sync.`,
      };
    } catch (error) {
      return toUnexpectedToolError(
        error,
        "Failed to create GitHub issue sync link",
      );
    }
  },
});

export const deleteGitHubIssueSyncLinkTool = tool({
  description:
    "Delete an existing GitHub issue sync link. Use getGitHubIntegrationTool first to find the link.",
  inputSchema: z.object({
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms deletion."),
    linkId: z.string().describe("Issue sync link ID."),
  }),
  execute: async ({ linkId, confirmed }, options) => {
    try {
      if (!confirmed) {
        return requireConfirmation("delete this GitHub issue sync link");
      }

      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const result = await deleteGitHubIssueSyncLinkAction(
        linkId,
        ctx.workspaceSlug,
      );

      if (result.error) {
        return {
          success: false,
          error: apiErrorMessage(result, "Failed to delete issue sync link"),
        };
      }

      return {
        success: true,
        message: "GitHub issue sync link deleted.",
      };
    } catch (error) {
      return toUnexpectedToolError(
        error,
        "Failed to delete GitHub issue sync link",
      );
    }
  },
});

export const updateGitHubWorkspaceSettingsTool = tool({
  description:
    "Update workspace-level GitHub settings such as branch format, magic word linking, assignee sync, label sync, PR body population, and close-on-commit behavior.",
  inputSchema: workspaceSettingsUpdateSchema.extend({
    confirmed: confirmationSchema,
  }),
  execute: async ({ confirmed, ...input }, options) => {
    try {
      if (!confirmed) {
        return requireConfirmation("update GitHub workspace settings");
      }

      const updates = toWorkspaceSettingsUpdate(input);
      if (!Object.keys(updates).length) {
        return {
          success: false,
          error: "At least one GitHub workspace setting is required.",
        };
      }

      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const result = await updateGitHubWorkspaceSettingsAction(
        updates,
        ctx.workspaceSlug,
      );

      if (result.error || !result.data) {
        return {
          success: false,
          error: apiErrorMessage(
            result,
            "Failed to update GitHub workspace settings",
          ),
        };
      }

      return {
        success: true,
        settings: result.data,
        message: "GitHub workspace settings updated.",
      };
    } catch (error) {
      return toUnexpectedToolError(
        error,
        "Failed to update GitHub workspace settings",
      );
    }
  },
});
