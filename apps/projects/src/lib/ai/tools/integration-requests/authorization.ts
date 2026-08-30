import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import type { AuthenticatedIntegrationRequestContext } from "./context";

export const getIntegrationRequestGuestAccessError = async (
  ctx: AuthenticatedIntegrationRequestContext,
  error: string,
) => {
  const workspace = await getWorkspace(ctx);

  return workspace.userRole === "guest" ? error : null;
};
