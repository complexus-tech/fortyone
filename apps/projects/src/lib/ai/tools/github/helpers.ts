import type { ToolExecutionOptions } from "ai";
import { auth } from "@/auth";
import { ApiError, get, type WorkspaceCtx } from "@/lib/http/fetch";
import type { ApiResponse } from "@/types/api-response";
import type {
  GitHubRepository,
  GitHubToolStory,
  GitHubToolTeam,
} from "./contracts";

type AuthenticatedContextResult = WorkspaceCtx | { error: string };

const getWorkspaceSlug = (experimentalContext: unknown) => {
  if (
    !experimentalContext ||
    typeof experimentalContext !== "object" ||
    !("workspaceSlug" in experimentalContext)
  ) {
    return undefined;
  }

  const workspaceSlug = experimentalContext.workspaceSlug;
  return typeof workspaceSlug === "string" ? workspaceSlug : undefined;
};

const getStoryAtPath = async (path: string, ctx: WorkspaceCtx) => {
  try {
    const response = await get<ApiResponse<GitHubToolStory>>(path, ctx);
    return response.data;
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return null;
    }

    throw error;
  }
};

export const getAuthenticatedGitHubContext = async ({
  experimental_context: experimentalContext,
}: ToolExecutionOptions): Promise<AuthenticatedContextResult> => {
  const session = await auth();

  if (!session) {
    return { error: "Authentication required to access GitHub integration" };
  }

  const workspaceSlug = getWorkspaceSlug(experimentalContext);
  if (!workspaceSlug) {
    return {
      error: "Workspace context is required to access GitHub integration",
    };
  }

  return { session, workspaceSlug };
};

export const apiErrorMessage = (
  result: { error?: { message?: string } },
  fallback: string,
) => result.error?.message || fallback;

export const toUnexpectedToolError = (error: unknown, fallback: string) => ({
  success: false,
  error: error instanceof Error ? error.message : fallback,
});

export const requireConfirmation = (action: string) => ({
  success: false,
  needsConfirmation: true,
  message: `Please confirm before I ${action}.`,
});

export const normalize = (value: string) => value.trim().toLowerCase();

export const resolveRepository = (
  repositories: GitHubRepository[],
  repositoryId?: string,
  repositoryFullName?: string,
) =>
  (repositoryId
    ? repositories.find((repository) => repository.id === repositoryId)
    : undefined) ??
  (repositoryFullName
    ? repositories.find(
        (repository) =>
          normalize(repository.fullName) === normalize(repositoryFullName),
      )
    : undefined);

export const getWorkspaceTeams = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<GitHubToolTeam[]>>("teams", ctx);
  return response.data!;
};

export const resolveTeam = (
  teams: GitHubToolTeam[],
  teamId?: string,
  teamName?: string,
) =>
  (teamId ? teams.find((team) => team.id === teamId) : undefined) ??
  (teamName
    ? teams.find((team) => normalize(team.name) === normalize(teamName))
    : undefined);

export const resolveStory = ({
  ctx,
  storyId,
  storyRef,
}: {
  ctx: WorkspaceCtx;
  storyId?: string;
  storyRef?: string;
}) => {
  if (storyId) {
    return getStoryAtPath(`stories/${storyId}`, ctx);
  }

  if (storyRef) {
    return getStoryAtPath(`story-by-ref/${storyRef}`, ctx);
  }

  return null;
};

export const getStoryDisplayRef = (story: GitHubToolStory) =>
  `${story.teamCode}-${story.sequenceId}`;
