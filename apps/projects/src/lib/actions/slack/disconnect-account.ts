import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const disconnectSlackAccountAction = async (workspaceSlug: string) => {
  try {
    const session = await auth();
    return await remove<ApiResponse<null>>("integrations/slack/account-link", {
      session: session!,
      workspaceSlug,
    });
  } catch (error) {
    return getApiError(error);
  }
};
