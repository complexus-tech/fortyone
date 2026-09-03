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
    const response = await post<{ url: string }, ApiResponse<StoryFigmaLink>>(
      `stories/${storyId}/figma-links`,
      { url },
      await context(workspaceSlug),
    );
    if (!response.data) {
      return { data: null, error: response.error } satisfies ApiResponse<null>;
    }
    if (response.data.unavailableAt) {
      const link: Link = {
        id: response.data.storyLinkId ?? response.data.id,
        storyId: response.data.storyId,
        url: response.data.artifact.canonicalUrl,
        title:
          title?.trim() ||
          response.data.artifact.nodeName ||
          response.data.artifact.fileName ||
          "Figma design",
        createdAt: response.data.createdAt,
        updatedAt: response.data.updatedAt,
      };
      return { data: { kind: "generic", link } };
    }
    return { data: { kind: "figma", link: response.data } };
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
