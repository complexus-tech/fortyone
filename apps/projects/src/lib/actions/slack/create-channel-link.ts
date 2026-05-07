import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  CreateSlackChannelLinkInput,
  SlackChannelLink,
} from "@/modules/settings/workspace/integrations/slack/types";

export const createSlackChannelLinkAction = async (
  input: CreateSlackChannelLinkInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      CreateSlackChannelLinkInput,
      ApiResponse<SlackChannelLink>
    >("integrations/slack/channel-links", input, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
