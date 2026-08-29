"use server";

import { auth } from "@/auth";
import { post, remove } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";
import { getApiError } from "@/utils";
import type {
  FigmaArtifact,
  StoryFigmaLink,
} from "@/modules/settings/workspace/integrations/figma/types";

const context = async (workspaceSlug: string) => ({
  session: (await auth())!,
  workspaceSlug,
});

export const createFigmaInstallSessionAction = async (
  workspaceSlug: string,
) => {
  try {
    return await post<
      Record<string, never>,
      ApiResponse<{ authorizationUrl: string }>
    >("integrations/figma/install-session", {}, await context(workspaceSlug));
  } catch (error) {
    return getApiError(error);
  }
};

export const disconnectFigmaAction = async (workspaceSlug: string) => {
  try {
    return await remove<ApiResponse<null>>(
      "integrations/figma",
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const resolveFigmaLinkAction = async (
  workspaceSlug: string,
  url: string,
) => {
  try {
    return await post<{ url: string }, ApiResponse<FigmaArtifact>>(
      "integrations/figma/resolve-link",
      { url },
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const linkFigmaStoryAction = async (
  workspaceSlug: string,
  storyId: string,
  url: string,
  title?: string,
): Promise<ApiResponse<LinkFigmaStoryResult>> => {
  try {
    const requestContext = await context(workspaceSlug);
    const response = await post<{ url: string }, ApiResponse<StoryFigmaLink>>(
      `stories/${storyId}/figma-links`,
      { url },
      requestContext,
    );
    const rateLimited =
      response.error?.code === "rate_limited" ||
      response.error?.message.startsWith("Figma rate limit reached");
    if (!rateLimited) {
      return response.data
        ? ({
            data: { kind: "figma", link: response.data },
          } satisfies ApiResponse<LinkFigmaStoryResult>)
        : ({ data: null, error: response.error } satisfies ApiResponse<null>);
    }

    const fallback = await post<
      { storyId: string; title: string; url: string },
      ApiResponse<Link>
    >(
      "links",
      {
        storyId,
        title: title?.trim() || "Figma design",
        url,
      },
      requestContext,
    );
    return fallback.data
      ? ({
          data: { kind: "generic", link: fallback.data },
        } satisfies ApiResponse<LinkFigmaStoryResult>)
      : ({ data: null, error: fallback.error } satisfies ApiResponse<null>);
  } catch (error) {
    const failure = getApiError(error);
    return { data: null, error: failure.error };
  }
};

export type LinkFigmaStoryResult =
  | { kind: "figma"; link: StoryFigmaLink }
  | { kind: "generic"; link: Link };

export const deleteFigmaStoryLinkAction = async (
  workspaceSlug: string,
  storyId: string,
  linkId: string,
) => {
  try {
    return await remove<ApiResponse<null>>(
      `stories/${storyId}/figma-links/${linkId}`,
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const refreshFigmaStoryLinkAction = async (
  workspaceSlug: string,
  storyId: string,
  linkId: string,
) => {
  try {
    return await post<Record<string, never>, ApiResponse<StoryFigmaLink>>(
      `stories/${storyId}/figma-links/${linkId}/refresh`,
      {},
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};
