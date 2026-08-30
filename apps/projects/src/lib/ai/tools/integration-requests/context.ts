import { auth } from "@/auth";
import type { WorkspaceCtx } from "@/lib/http/fetch";

export type AuthenticatedIntegrationRequestContext = WorkspaceCtx & {
  session: NonNullable<WorkspaceCtx["session"]>;
};

export type IntegrationRequestToolContextResult =
  | AuthenticatedIntegrationRequestContext
  | { error: string };

export const getAuthenticatedIntegrationRequestContext = async (
  experimentalContext: unknown,
): Promise<IntegrationRequestToolContextResult> => {
  const session = await auth();

  if (!session) {
    return { error: "Authentication required to access integration requests" };
  }

  const workspaceSlug = (experimentalContext as { workspaceSlug?: string })
    .workspaceSlug;

  if (!workspaceSlug) {
    return { error: "Workspace context is required" };
  }

  return { session, workspaceSlug };
};
