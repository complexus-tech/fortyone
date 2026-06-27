import { getApiUrl } from "@/lib/api-url";
import { applyPublicPortalWorkspace } from "./data";
import { publicPortalFixture } from "./fixtures";
import type { PublicPortal, PublicPortalWorkspace } from "./types";

type ApiResponse<T> = {
  data: T;
};

export const getPublicPortal = async (
  portalSlug: string,
): Promise<PublicPortal> => {
  const apiUrl = getApiUrl();

  if (!apiUrl) {
    return publicPortalFixture;
  }

  try {
    const response = await fetch(`${apiUrl}/portals/${portalSlug}`, {
      cache: "no-store",
    });

    if (!response.ok) {
      return publicPortalFixture;
    }

    const payload =
      (await response.json()) as ApiResponse<PublicPortalWorkspace>;

    return applyPublicPortalWorkspace(publicPortalFixture, payload.data);
  } catch {
    return publicPortalFixture;
  }
};
