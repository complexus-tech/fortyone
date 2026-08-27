import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import type { ApiResponse } from "@/types";
import { getLatestAiChatAssistantMessage } from "../queries/get-latest-ai-chat-assistant-message";

export type MutationApprovalExecution = {
  state:
    | "claimed"
    | "completed"
    | "executing"
    | "failed_uncertain"
    | "in_progress"
    | "ready"
    | "started";
  failureCode?: string;
  leaseExpiresAt?: string;
  leaseToken?: string;
  output?: unknown;
};

export type MutationApprovalFailureCode =
  | "completion_persistence_uncertain"
  | "start_transition_uncertain";

type MutationApprovalIdentity = {
  chatId: string;
  fingerprint: string;
  toolCallId: string;
  workspaceSlug: string;
};

type LeasedMutationApprovalIdentity = MutationApprovalIdentity & {
  leaseToken: string;
};

const getAuthenticatedWorkspaceContext = async (workspaceSlug: string) => {
  const session = await auth();
  if (!session?.user) {
    throw new Error("Mutation approval authentication is required.");
  }
  if (!workspaceSlug) {
    throw new Error("Mutation approval workspace is required.");
  }

  return { session, workspaceSlug };
};

const getMutationApprovalPath = (chatId: string, toolCallId: string) =>
  `chat-sessions/${encodeURIComponent(chatId)}/mutation-approvals/${encodeURIComponent(toolCallId)}`;

export const claimMutationApprovalExecution = async ({
  chatId,
  fingerprint,
  toolCallId,
  workspaceSlug,
}: MutationApprovalIdentity): Promise<MutationApprovalExecution> => {
  const ctx = await getAuthenticatedWorkspaceContext(workspaceSlug);
  const response = await post<
    { fingerprint: string },
    ApiResponse<MutationApprovalExecution>
  >(
    `${getMutationApprovalPath(chatId, toolCallId)}/claim`,
    { fingerprint },
    ctx,
  );
  if (!response.data) {
    throw new Error("Mutation approval claim returned no execution state.");
  }
  if (response.data.state === "claimed" && !response.data.leaseToken) {
    throw new Error("Mutation approval claim returned no execution lease.");
  }

  return response.data;
};

export const startMutationApprovalExecution = async ({
  chatId,
  fingerprint,
  leaseToken,
  toolCallId,
  workspaceSlug,
}: LeasedMutationApprovalIdentity): Promise<MutationApprovalExecution> => {
  const ctx = await getAuthenticatedWorkspaceContext(workspaceSlug);
  const response = await post<
    { fingerprint: string; leaseToken: string },
    ApiResponse<MutationApprovalExecution>
  >(
    `${getMutationApprovalPath(chatId, toolCallId)}/start`,
    { fingerprint, leaseToken },
    ctx,
  );
  if (!response.data) {
    throw new Error("Mutation approval start returned no execution state.");
  }

  return response.data;
};

export const completeMutationApprovalExecution = async ({
  chatId,
  fingerprint,
  leaseToken,
  output,
  toolCallId,
  workspaceSlug,
}: LeasedMutationApprovalIdentity & {
  output: unknown;
}): Promise<MutationApprovalExecution> => {
  const ctx = await getAuthenticatedWorkspaceContext(workspaceSlug);
  const response = await post<
    { fingerprint: string; leaseToken: string; output: unknown },
    ApiResponse<MutationApprovalExecution>
  >(
    `${getMutationApprovalPath(chatId, toolCallId)}/complete`,
    { fingerprint, leaseToken, output: output ?? null },
    ctx,
  );
  if (!response.data) {
    throw new Error(
      "Mutation approval completion returned no execution state.",
    );
  }

  return response.data;
};

export const failMutationApprovalExecution = async ({
  chatId,
  failureCode,
  fingerprint,
  leaseToken,
  toolCallId,
  workspaceSlug,
}: LeasedMutationApprovalIdentity & {
  failureCode: MutationApprovalFailureCode;
}): Promise<MutationApprovalExecution> => {
  const ctx = await getAuthenticatedWorkspaceContext(workspaceSlug);
  const response = await post<
    {
      failureCode: MutationApprovalFailureCode;
      fingerprint: string;
      leaseToken: string;
    },
    ApiResponse<MutationApprovalExecution>
  >(
    `${getMutationApprovalPath(chatId, toolCallId)}/fail`,
    { failureCode, fingerprint, leaseToken },
    ctx,
  );
  if (!response.data) {
    throw new Error("Mutation approval failure returned no execution state.");
  }

  return response.data;
};

export const getPersistedMutationApprovalMessage = async (
  chatId: string,
  workspaceSlug: string,
): Promise<MayaUIMessage | null> => {
  const ctx = await getAuthenticatedWorkspaceContext(workspaceSlug);
  return getLatestAiChatAssistantMessage(ctx, chatId);
};
