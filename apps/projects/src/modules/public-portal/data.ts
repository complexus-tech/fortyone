import type { PublicPortal, PublicPortalWorkspace } from "./types";

export const applyPublicPortalWorkspace = (
  portal: PublicPortal,
  workspace: PublicPortalWorkspace,
): PublicPortal => ({
  ...portal,
  name: workspace.name,
  slug: workspace.slug,
  workspace,
});
