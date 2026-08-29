import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type LinkSlackAccountInput = {
  token: string;
};

export type LinkSlackAccountResult = {
  status: "connected" | "already_connected";
  slackUserId: string;
};

export const linkSlackAccountAction = async (
  workspaceSlug: string,
  input: LinkSlackAccountInput,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      LinkSlackAccountInput,
      ApiResponse<LinkSlackAccountResult>
    >("integrations/slack/link-account", input, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
