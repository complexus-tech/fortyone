import "server-only";
import { ApiError } from "api-client";
import type { Session } from "@/auth";
import type { ApiResponse, Subscription } from "@/types";
import { get } from "@/lib/http";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getMemories } from "@/modules/ai-chats/queries/get-memory";
import type { ChatRequestBody } from "./chat-request";
import {
  ChatAdmissionError,
  resolveChatMessageLimit,
  resolveChatTimezone,
} from "./request-context-policy";

export const resolveTrustedChatContext = async (
  request: ChatRequestBody,
  session: Session,
) => {
  const slug = request.workspace?.slug;
  if (
    typeof slug !== "string" ||
    !/^[a-z0-9][a-z0-9-]{0,254}$/.test(slug) ||
    typeof request.id !== "string" ||
    request.id.length !== 16
  ) {
    throw new ChatAdmissionError(
      "A valid conversation and workspace are required.",
      400,
    );
  }
  const timezone = resolveChatTimezone(request.timezone);
  const ctx = { session, workspaceSlug: slug };
  const workspace = await getWorkspace(ctx);
  if (!workspace.id || workspace.deletedAt)
    throw new ChatAdmissionError("Workspace access is unavailable.", 403);
  const [settings, memories, subscription] = await Promise.all([
    getWorkspaceSettings(ctx),
    getMemories(ctx),
    get<ApiResponse<Subscription>>("subscription", ctx)
      .then((result) => result.data ?? null)
      .catch((error: unknown) => {
        if (error instanceof ApiError && error.status === 404) return null;
        throw error;
      }),
  ]);
  const plural = (term: string) =>
    term.endsWith("y") ? `${term.slice(0, -1)}ies` : `${term}s`;
  return {
    workspace,
    memories,
    subscription,
    timezone,
    username: session.user.username,
    terminology: {
      stories: plural(settings.storyTerm),
      sprints: plural(settings.sprintTerm),
      objectives: plural(settings.objectiveTerm),
      keyResults: plural(settings.keyResultTerm),
    },
    messageLimit: resolveChatMessageLimit(workspace, subscription),
  };
};
