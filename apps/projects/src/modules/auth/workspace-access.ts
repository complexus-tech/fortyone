import { z } from "zod";
import type { WorkspaceCtx } from "@/lib/http";
import { ApiError } from "@/lib/http";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

const ssoDenialSchema = z.object({
  error: z.object({ code: z.literal("workspace_sso_required") }),
});

// Only the authoritative scoped denial can select SSO recovery. Public provider
// configuration cannot classify other session policies or grant tenant access.
export const getWorkspaceAccess = async (
  ctx: WorkspaceCtx,
): Promise<"allowed" | "sso-required" | "denied"> => {
  try {
    await getWorkspace(ctx);
    return "allowed";
  } catch (error) {
    if (!(error instanceof ApiError)) throw error;
    if (error.status === 403 && ssoDenialSchema.safeParse(error.data).success) {
      return "sso-required";
    }
    if (error.status === 404) return "denied";
    throw error;
  }
};
