import { tool } from "ai";
import { z } from "zod";
import { auth } from "@/auth";
import type { WorkspaceCtx } from "@/lib/http";
import { createGitHubInstallSessionAction } from "@/lib/actions/github/create-install-session";
import { createGitHubIssueSyncLinkAction } from "@/lib/actions/github/create-issue-sync-link";
import { deleteGitHubIssueSyncLinkAction } from "@/lib/actions/github/delete-issue-sync-link";
import { deleteStoryGitHubLinkAction } from "@/lib/actions/github/delete-story-github-link";
import { postGitHubCommentAction } from "@/lib/actions/github/post-github-comment";
import { resyncGitHubRepositoriesAction } from "@/lib/actions/github/resync-repositories";
import { updateGitHubTeamSettingsAction } from "@/lib/actions/github/update-team-settings";
import { updateGitHubWorkspaceSettingsAction } from "@/lib/actions/github/update-workspace-settings";
import { getGitHubIntegration } from "@/lib/queries/github/get-integration";
import { getStoryGitHubComments } from "@/lib/queries/github/get-story-github-comments";
import { getStoryGitHubLinks } from "@/lib/queries/github/get-story-github-links";
import { getGitHubTeamSettings } from "@/lib/queries/github/get-team-settings";
import { getStory, getStoryRef } from "@/modules/story/queries/get-story";
import { getTeams } from "@/modules/teams/queries/get-teams";
import type {
  GitHubRepository,
  GitHubWorkflowRule,
  StoryGitHubLink,
} from "@/modules/settings/workspace/integrations/github/types";

const syncDirectionSchema = z.enum(["inbound_only", "bidirectional"]);

const teamRuleSchema = z.object({
  eventKey: z
    .string()
    .describe(
      "GitHub workflow event key, for example issue.opened, issue.closed, pull_request.opened.",
    ),
  targetStatusId: z
    .string()
    .nullable()
    .optional()
    .describe("Target FortyOne status ID. Resolve statuses first."),
  baseBranchPattern: z
    .string()
    .nullable()
    .optional()
    .describe("Optional base branch pattern for pull request rules."),
  isActive: z.boolean().describe("Whether the automation rule is active."),
});

const getAuthenticatedContext = async (
  experimentalContext: unknown,
): Promise<WorkspaceCtx | { error: string }> => {
  const session = await auth();

  if (!session) {
    return { error: "Authentication required to access GitHub integration" };
  }

  const workspaceSlug = (experimentalContext as { workspaceSlug?: string })
    .workspaceSlug;

  if (!workspaceSlug) {
    return {
      error: "Workspace context is required to access GitHub integration",
    };
  }

  return { session, workspaceSlug };
};

const apiErrorMessage = (
  result: { error?: { message?: string } },
  fallback: string,
) => result.error?.message || fallback;

const normalize = (value: string) => value.trim().toLowerCase();

const toRepositorySummary = (repository: GitHubRepository) => ({
  id: repository.id,
  fullName: repository.fullName,
  ownerLogin: repository.ownerLogin,
  name: repository.name,
  htmlUrl: repository.htmlUrl,
  defaultBranch: repository.defaultBranch,
  isPrivate: repository.isPrivate,
  isArchived: repository.isArchived,
  isDisabled: repository.isDisabled,
  isActive: repository.isActive,
  lastSyncedAt: repository.lastSyncedAt,
});

const toStoryLinkSummary = (link: StoryGitHubLink) => ({
  id: link.id,
  type: link.externalType,
  number: link.githubNumber,
  title: link.title,
  state: link.state,
  reviewState: link.reviewState,
  checkState: link.checkState,
  repositoryFullName: link.repositoryFullName,
  refName: link.refName,
  url: link.url,
  createdAt: link.createdAt,
});

const resolveStory = async ({
  ctx,
  storyId,
  storyRef,
}: {
  ctx: WorkspaceCtx;
  storyId?: string;
  storyRef?: string;
}) => {
  if (storyId) {
    return getStory(storyId, ctx);
  }

  if (storyRef) {
    return getStoryRef(storyRef, ctx);
  }

  return null;
};

const getStoryDisplayRef = (story: { teamCode: string; sequenceId: number }) =>
  `${story.teamCode}-${story.sequenceId}`;

const requireConfirmation = (action: string) => ({
  success: false,
  needsConfirmation: true,
  message: `Please confirm before I ${action}.`,
});

export const getGitHubIntegrationTool = tool({
  description:
    "Get the workspace GitHub integration status, installations, repositories, issue sync links, and workspace GitHub settings.",
  inputSchema: z.object({}),
  execute: async ({}, { experimental_context }) => {
    try {
      const ctx = await getAuthenticatedContext(experimental_context);
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
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to get GitHub integration",
      };
    }
  },
});

export const createGitHubInstallSessionTool = tool({
  description:
    "Create a GitHub App installation session for the current workspace. Use this when GitHub is not connected and the user wants to connect it.",
  inputSchema: z.object({}),
  execute: async ({}, { experimental_context }) => {
    try {
      const ctx = await getAuthenticatedContext(experimental_context);
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
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to create GitHub install session",
      };
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
  execute: async ({ confirmed }, { experimental_context }) => {
    try {
      if (!confirmed) {
        return requireConfirmation("resync GitHub repositories");
      }

      const ctx = await getAuthenticatedContext(experimental_context);
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
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to resync GitHub repositories",
      };
    }
  },
});

export const createGitHubIssueSyncLinkTool = tool({
  description:
    "Create an issue sync link between a GitHub repository and a FortyOne team. Resolve repository and team by ID or exact name.",
  inputSchema: z.object({
    repositoryId: z.string().optional().describe("GitHub repository ID."),
    repositoryFullName: z
      .string()
      .optional()
      .describe("GitHub repository full name, for example owner/repo."),
    teamId: z.string().optional().describe("FortyOne team ID."),
    teamName: z.string().optional().describe("FortyOne team name."),
    syncDirection: syncDirectionSchema.default("inbound_only"),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms linking."),
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
    { experimental_context },
  ) => {
    try {
      if (!confirmed) {
        return requireConfirmation("link this GitHub repository to the team");
      }

      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const [integration, teams] = await Promise.all([
        getGitHubIntegration(ctx),
        getTeams(ctx),
      ]);

      const repository =
        (repositoryId
          ? integration.repositories.find((repo) => repo.id === repositoryId)
          : undefined) ??
        (repositoryFullName
          ? integration.repositories.find(
              (repo) =>
                normalize(repo.fullName) === normalize(repositoryFullName),
            )
          : undefined);

      if (!repository) {
        return {
          success: false,
          error:
            "GitHub repository not found. Ask the user for the exact repository or resync repositories first.",
        };
      }

      const team =
        (teamId ? teams.find((item) => item.id === teamId) : undefined) ??
        (teamName
          ? teams.find((item) => normalize(item.name) === normalize(teamName))
          : undefined);

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
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to create GitHub issue sync link",
      };
    }
  },
});

export const deleteGitHubIssueSyncLinkTool = tool({
  description:
    "Delete an existing GitHub issue sync link. Use getGitHubIntegrationTool first to find the link.",
  inputSchema: z.object({
    linkId: z.string().describe("Issue sync link ID."),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms deletion."),
  }),
  execute: async ({ linkId, confirmed }, { experimental_context }) => {
    try {
      if (!confirmed) {
        return requireConfirmation("delete this GitHub issue sync link");
      }

      const ctx = await getAuthenticatedContext(experimental_context);
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
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to delete GitHub issue sync link",
      };
    }
  },
});

export const updateGitHubWorkspaceSettingsTool = tool({
  description:
    "Update workspace-level GitHub settings such as branch format, magic word linking, assignee sync, label sync, PR body population, and close-on-commit behavior.",
  inputSchema: z.object({
    branchFormat: z.string().optional(),
    linkCommitsByMagicWords: z.boolean().optional(),
    syncAssignees: z.boolean().optional(),
    syncLabels: z.boolean().optional(),
    autoPopulatePrBody: z.boolean().optional(),
    closeOnCommitKeywords: z.boolean().optional(),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms changes."),
  }),
  execute: async ({ confirmed, ...input }, { experimental_context }) => {
    try {
      if (!confirmed) {
        return requireConfirmation("update GitHub workspace settings");
      }

      const updates = Object.fromEntries(
        Object.entries(input).filter(([, value]) => value !== undefined),
      );

      if (!Object.keys(updates).length) {
        return {
          success: false,
          error: "At least one GitHub workspace setting is required.",
        };
      }

      const ctx = await getAuthenticatedContext(experimental_context);
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
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to update GitHub workspace settings",
      };
    }
  },
});

export const getGitHubTeamSettingsTool = tool({
  description:
    "Get GitHub automation settings and workflow rules for a FortyOne team.",
  inputSchema: z.object({
    teamId: z.string().optional().describe("FortyOne team ID."),
    teamName: z.string().optional().describe("FortyOne team name."),
  }),
  execute: async ({ teamId, teamName }, { experimental_context }) => {
    try {
      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const teams = await getTeams(ctx);
      const team =
        (teamId ? teams.find((item) => item.id === teamId) : undefined) ??
        (teamName
          ? teams.find((item) => normalize(item.name) === normalize(teamName))
          : undefined);

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
        rules: settings.rules.map((rule: GitHubWorkflowRule) => ({
          id: rule.id,
          eventKey: rule.eventKey,
          targetStatusId: rule.targetStatusId,
          baseBranchPattern: rule.baseBranchPattern,
          isActive: rule.isActive,
        })),
        message: `${team.name} has ${settings.rules.length} GitHub automation rule${settings.rules.length === 1 ? "" : "s"}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to get GitHub team settings",
      };
    }
  },
});

export const updateGitHubTeamSettingsTool = tool({
  description:
    "Replace GitHub automation rules for a FortyOne team. Read existing settings first and send the complete desired rules array.",
  inputSchema: z.object({
    teamId: z.string().optional().describe("FortyOne team ID."),
    teamName: z.string().optional().describe("FortyOne team name."),
    rules: z.array(teamRuleSchema),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms changes."),
  }),
  execute: async (
    { teamId, teamName, rules, confirmed },
    { experimental_context },
  ) => {
    try {
      if (!confirmed) {
        return requireConfirmation(
          "update this team's GitHub automation rules",
        );
      }

      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const teams = await getTeams(ctx);
      const team =
        (teamId ? teams.find((item) => item.id === teamId) : undefined) ??
        (teamName
          ? teams.find((item) => normalize(item.name) === normalize(teamName))
          : undefined);

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
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to update GitHub team settings",
      };
    }
  },
});

export const getStoryGitHubLinksTool = tool({
  description:
    "Get GitHub issues, pull requests, branches, or commits linked to a FortyOne story.",
  inputSchema: z.object({
    storyId: z.string().optional().describe("Story ID."),
    storyRef: z
      .string()
      .optional()
      .describe("Story reference such as WEB-123 if story ID is unknown."),
  }),
  execute: async ({ storyId, storyRef }, { experimental_context }) => {
    try {
      if (!storyId && !storyRef) {
        return {
          success: false,
          error: "Either storyId or storyRef is required.",
        };
      }

      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const story = await resolveStory({ ctx, storyId, storyRef });

      if (!story) {
        return {
          success: false,
          error: "Story not found.",
        };
      }

      const links = await getStoryGitHubLinks(story.id, ctx);
      const displayRef = getStoryDisplayRef(story);

      return {
        success: true,
        kind: "github-story-report",
        title: `${displayRef} GitHub links`,
        story: {
          id: story.id,
          ref: displayRef,
          title: story.title,
        },
        links: links.map(toStoryLinkSummary),
        message: `${displayRef} has ${links.length} GitHub link${links.length === 1 ? "" : "s"}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to get story GitHub links",
      };
    }
  },
});

export const getStoryGitHubCommentsTool = tool({
  description:
    "Get comments from the GitHub issue or pull request linked to a FortyOne story.",
  inputSchema: z.object({
    storyId: z.string().optional().describe("Story ID."),
    storyRef: z
      .string()
      .optional()
      .describe("Story reference such as WEB-123 if story ID is unknown."),
  }),
  execute: async ({ storyId, storyRef }, { experimental_context }) => {
    try {
      if (!storyId && !storyRef) {
        return {
          success: false,
          error: "Either storyId or storyRef is required.",
        };
      }

      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const story = await resolveStory({ ctx, storyId, storyRef });

      if (!story) {
        return {
          success: false,
          error: "Story not found.",
        };
      }

      const comments = await getStoryGitHubComments(story.id, ctx);
      const displayRef = getStoryDisplayRef(story);

      return {
        success: true,
        story: {
          id: story.id,
          ref: displayRef,
          title: story.title,
        },
        comments: comments.map((comment) => ({
          id: comment.id,
          body: comment.body,
          userLogin: comment.userLogin,
          userAvatar: comment.userAvatar,
          createdAt: comment.createdAt,
          updatedAt: comment.updatedAt,
          htmlUrl: comment.htmlUrl,
        })),
        message: `Found ${comments.length} GitHub comment${comments.length === 1 ? "" : "s"} for ${displayRef}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to get story GitHub comments",
      };
    }
  },
});

export const postStoryGitHubCommentTool = tool({
  description:
    "Post a comment to the GitHub issue or pull request linked to a FortyOne story.",
  inputSchema: z.object({
    storyId: z.string().optional().describe("Story ID."),
    storyRef: z
      .string()
      .optional()
      .describe("Story reference such as WEB-123 if story ID is unknown."),
    body: z.string().min(1).describe("Comment body to post to GitHub."),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms posting."),
  }),
  execute: async (
    { storyId, storyRef, body, confirmed },
    { experimental_context },
  ) => {
    try {
      if (!confirmed) {
        return requireConfirmation("post this comment to GitHub");
      }

      if (!storyId && !storyRef) {
        return {
          success: false,
          error: "Either storyId or storyRef is required.",
        };
      }

      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const story = await resolveStory({ ctx, storyId, storyRef });

      if (!story) {
        return {
          success: false,
          error: "Story not found.",
        };
      }

      const result = await postGitHubCommentAction(
        story.id,
        { body },
        ctx.workspaceSlug,
      );
      const displayRef = getStoryDisplayRef(story);

      if (result.error) {
        return {
          success: false,
          error: apiErrorMessage(result, "Failed to post GitHub comment"),
        };
      }

      return {
        success: true,
        message: `Posted the comment to GitHub for ${displayRef}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to post GitHub comment",
      };
    }
  },
});

export const deleteStoryGitHubLinkTool = tool({
  description:
    "Remove a GitHub link from a FortyOne story. Use getStoryGitHubLinksTool first to find the link.",
  inputSchema: z.object({
    storyId: z.string().optional().describe("Story ID."),
    storyRef: z
      .string()
      .optional()
      .describe("Story reference such as WEB-123 if story ID is unknown."),
    linkId: z.string().describe("Story GitHub link ID."),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms unlinking."),
  }),
  execute: async (
    { storyId, storyRef, linkId, confirmed },
    { experimental_context },
  ) => {
    try {
      if (!confirmed) {
        return requireConfirmation("remove this GitHub link from the story");
      }

      if (!storyId && !storyRef) {
        return {
          success: false,
          error: "Either storyId or storyRef is required.",
        };
      }

      const ctx = await getAuthenticatedContext(experimental_context);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const story = await resolveStory({ ctx, storyId, storyRef });

      if (!story) {
        return {
          success: false,
          error: "Story not found.",
        };
      }

      const result = await deleteStoryGitHubLinkAction(
        story.id,
        linkId,
        ctx.workspaceSlug,
      );
      const displayRef = getStoryDisplayRef(story);

      if (result.error) {
        return {
          success: false,
          error: apiErrorMessage(result, "Failed to remove GitHub link"),
        };
      }

      return {
        success: true,
        message: `Removed the GitHub link from ${displayRef}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to remove story GitHub link",
      };
    }
  },
});
