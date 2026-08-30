import "server-only";

import { createUIMessageStream, createUIMessageStreamResponse } from "ai";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { recoverMutationApprovalOutput } from "@/modules/ai-chats/actions/message-write";
import { getChatErrorDiagnostic } from "../chat-errors";
import { beginChatWrite, saveChat } from "../save-chat";
import { getMutationToolApprovals } from "./approval-policy";
import { executeApprovalOnce } from "./execution-ledger";

const SKIPPED_APPROVAL_OUTPUT_MESSAGE =
  "Maya did not run this approved change because an earlier approved change was unresolved. Review the earlier result, then ask Maya to prepare this change again.";

const getSkippedApprovalOutput = () => ({
  error: SKIPPED_APPROVAL_OUTPUT_MESSAGE,
  success: false,
});

export const createMutationToolApprovalResponse = ({
  abortSignal,
  chatId,
  messageId,
  messages,
  userId,
  workspaceSlug,
}: {
  abortSignal: AbortSignal;
  chatId: string;
  messageId?: string;
  messages: MayaUIMessage[];
  userId?: string;
  workspaceSlug: string;
}): Promise<Response> | Response | undefined => {
  const submittedApprovals = getMutationToolApprovals(messages);
  if (submittedApprovals.length === 0) return undefined;
  if (!userId) return new Response("Unauthorized", { status: 401 });

  return (async () => {
    // Reserve and validate the persisted transition before creating the
    // persistence stream. If the server repaired durable receipts, its
    // request-safe transcript becomes the base for both streaming and final
    // CAS persistence instead of letting a stale browser overwrite it.
    const reservation = await beginChatWrite({
      id: chatId,
      messageId,
      messages,
      operation: "approval",
      workspaceSlug,
    });
    const canonicalMessages = reservation.messages ?? messages;
    const approvals = getMutationToolApprovals(canonicalMessages);
    const recoverableApprovals = new Map<string, string>();

    const stream = createUIMessageStream<MayaUIMessage>({
      execute: async ({ writer }) => {
        // The server may have repaired terminal receipts while reserving this
        // write. Re-emit them so the browser converges immediately instead of
        // waiting for a refresh; replaying an identical terminal chunk is
        // idempotent in AI SDK's UI-message state machine.
        const canonicalLastMessage = canonicalMessages.at(-1);
        if (canonicalLastMessage?.role === "assistant") {
          for (const part of canonicalLastMessage.parts) {
            if (
              !part.type.startsWith("tool-") ||
              !("state" in part) ||
              !("toolCallId" in part) ||
              typeof part.toolCallId !== "string"
            ) {
              continue;
            }
            if (part.state === "output-denied") {
              writer.write({
                toolCallId: part.toolCallId,
                type: "tool-output-denied",
              });
            } else if (part.state === "output-available" && "output" in part) {
              writer.write({
                output: part.output,
                toolCallId: part.toolCallId,
                type: "tool-output-available",
              });
            }
          }
        }

        let haltFollowingApprovedMutations = false;
        for (const approval of approvals) {
          if (!approval.approved) {
            writer.write({
              toolCallId: approval.toolCallId,
              type: "tool-output-denied",
            });
            continue;
          }

          if (haltFollowingApprovedMutations) {
            writer.write({
              output: getSkippedApprovalOutput(),
              toolCallId: approval.toolCallId,
              type: "tool-output-available",
            });
            continue;
          }

          // eslint-disable-next-line no-await-in-loop -- Mutations must preserve the approved order and never execute concurrently.
          const result = await executeApprovalOnce({
            abortSignal,
            approval,
            chatId,
            userId,
            workspaceSlug,
          });

          if (result.denied) {
            writer.write({
              toolCallId: approval.toolCallId,
              type: "tool-output-denied",
            });
            continue;
          }

          if (result.durableFingerprint) {
            recoverableApprovals.set(
              approval.toolCallId,
              result.durableFingerprint,
            );
          }
          haltFollowingApprovedMutations = result.haltsFollowing;

          writer.write({
            output: result.output,
            toolCallId: approval.toolCallId,
            type: "tool-output-available",
          });
        }
      },
      onFinish: async ({ messages: finishedMessages }) => {
        let finalizationError: unknown;
        let applied = false;
        try {
          const result = await saveChat({
            id: chatId,
            messages: finishedMessages,
            reservation,
            workspaceSlug,
          });
          applied = result.applied;
        } catch (error) {
          finalizationError = error;
        }
        if (applied) return;

        if (recoverableApprovals.size === 0) {
          if (finalizationError) {
            throw finalizationError instanceof Error
              ? finalizationError
              : new Error("Maya transcript finalization failed.", {
                  cause: finalizationError,
                });
          }
          return;
        }

        try {
          for (const [toolCallId, fingerprint] of recoverableApprovals) {
            // eslint-disable-next-line no-await-in-loop -- Each targeted merge preserves the exact durable approval order and surrounding transcript.
            await recoverMutationApprovalOutput({
              chatId,
              fingerprint,
              toolCallId,
              workspaceSlug,
            });
          }
        } catch (recoveryError) {
          // eslint-disable-next-line no-console -- Recovery diagnostics omit message and tool payloads.
          console.error(
            "[chat/route] Could not project durable approval output into the transcript",
            getChatErrorDiagnostic(recoveryError),
          );
          throw finalizationError ?? recoveryError;
        }
      },
      originalMessages: canonicalMessages,
    });

    return createUIMessageStreamResponse({ stream });
  })();
};
