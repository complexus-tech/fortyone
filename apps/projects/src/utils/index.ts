import { ApiError } from "@/lib/http/error";
import type { ApiResponse, Workspace } from "@/types";
import type { Invitation } from "@/modules/invitations/types";

export const getRedirectUrl = (
  workspaces: Workspace[],
  invitations: Invitation[] = [],
  lastUsedWorkspaceId?: string,
) => {
  if (workspaces.length === 0) {
    if (invitations.length > 0) {
      return `/onboarding/join?token=${invitations[0].token}`;
    }
    return "/onboarding/create";
  }
  const activeWorkspace =
    workspaces.find((workspace) => workspace.id === lastUsedWorkspaceId) ||
    workspaces[0];
  return `/${activeWorkspace.slug}/my-work`;
};

export const buildWorkspaceUrl = (slug: string) => {
  return `/${slug}/my-work`;
};

export const withWorkspacePath = (path: string, slug?: string) => {
  if (!slug) {
    return path;
  }

  if (path.startsWith(`/${slug}`)) {
    return path;
  }

  if (path.startsWith("/")) {
    return `/${slug}${path}`;
  }

  return `/${slug}/${path}`;
};

export const slugify = (text = "") => {
  return text
    .toString()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .trim()
    .replace(/\s+/g, "-")
    .replace(/&/g, "-and-")
    .replace(/[^\w-]+/g, "")
    .replace(/--+/g, "-");
};

export const getApiError = (error: unknown): ApiResponse<null> => {
  if (error instanceof ApiError) {
    return error.data as ApiResponse<null>;
  }
  return {
    data: null,
    error: {
      message: "An error occurred",
    },
  };
};

/**
 * Alternative function that returns rgba format
 * @param hex - Hex color (with or without #)
 * @param opacity - Opacity value between 0 and 1
 * @returns rgba color string
 */
export const hexToRgba = (hex = "#6B665C", opacity = 0.1): string => {
  const cleanHex = hex.replace("#", "");

  if (!/^[0-9A-F]{6}$/i.test(cleanHex)) {
    throw new Error("Invalid hex color format");
  }

  if (opacity < 0 || opacity > 1) {
    throw new Error("Opacity must be between 0 and 1");
  }

  const r = parseInt(cleanHex.substring(0, 2), 16);
  const g = parseInt(cleanHex.substring(2, 4), 16);
  const b = parseInt(cleanHex.substring(4, 6), 16);

  return `rgba(${r}, ${g}, ${b}, ${opacity})`;
};

export const toTitleCase = (text: string) => {
  return text.charAt(0).toUpperCase() + text.slice(1);
};
