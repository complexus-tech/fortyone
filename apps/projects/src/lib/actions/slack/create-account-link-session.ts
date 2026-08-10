import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export type SlackAccountLinkSession = {
  linked: boolean;
  canLink: boolean;
  installUrl?: string;
};

export const createSlackAccountLinkSessionAction = async (
  workspaceSlug: string,
  returnUrl: string,
) => {
  try {
    const session = await auth();
    return await post<
      { returnUrl: string },
      ApiResponse<SlackAccountLinkSession>
    >(
      "integrations/slack/account-link-session",
      { returnUrl },
      { session: session!, workspaceSlug },
    );
  } catch (error) {
    return getApiError(error);
  }
};
