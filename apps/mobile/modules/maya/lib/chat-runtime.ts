import { Chat } from "@ai-sdk/react";
import type { FileUIPart } from "ai";
import {
  DefaultChatTransport,
  lastAssistantMessageIsCompleteWithApprovalResponses,
  safeValidateUIMessages,
} from "ai";
import type { MayaMessage } from "../types";
import type { createMayaCloudClient } from "./cloud-client";
import type { createMayaSessionGuard } from "./session-scope";
import type { VoiceHistoryMessage } from "./voice-history-queue";
import { validateMayaFileParts } from "./attachments";
import {
  asRecord,
  getPendingMayaApprovals,
  hasUnresolvedMayaApprovals,
  prepareMobileMayaMessages,
} from "./chat-protocol";

type RequestBody =
  | Record<string, unknown>
  | ((signal: AbortSignal) => Promise<Record<string, unknown>>);

type RuntimeOptions = {
  id: string;
  messages: MayaMessage[];
  persisted?: boolean;
  workspace: string;
  cloud: ReturnType<typeof createMayaCloudClient>;
  guard: ReturnType<typeof createMayaSessionGuard>;
  getHistory: (signal: AbortSignal) => Promise<MayaMessage[]>;
  onFinish: (message: MayaMessage, interrupted: boolean) => void;
};

/** Owns admission through the complete stream, including explicit approval continuations. */
export const createMayaChatRuntime = (options: RuntimeOptions) => {
  const lifetime = new AbortController();
  let pending: Promise<void> | null = null;
  let needsRecovery = false;
  let cancellationVersion = 0;
  let attachmentVersion = 0;
  let historyRequest: AbortController | null = null;
  let contextRequest: AbortController | null = null;
  let persisted = options.persisted === true || options.messages.length > 0;
  const chat = new Chat<MayaMessage>({
    id: options.id,
    messages: options.messages,
    transport: new DefaultChatTransport({
      api: options.cloud.url("/api/chat"),
      fetch: async (input, init) => {
        const response = await options.cloud.fetch(input, init);
        // A successful stream starts only after the server reserves history.
        persisted = true;
        return response;
      },
      prepareSendMessagesRequest: ({
        id,
        messages,
        trigger,
        messageId,
        body,
      }) => ({
        body: {
          ...body,
          client: "mobile",
          workspace: { ...asRecord(body?.workspace), slug: options.workspace },
          id,
          messages: prepareMobileMayaMessages(messages, {
            preserveLastUserFiles:
              trigger === "submit-message" && messages.at(-1)?.role === "user",
          }),
          trigger,
          messageId,
        },
      }),
    }),
    onFinish: ({ message, isAbort, isError, isDisconnect }) => {
      needsRecovery = isAbort || isError || isDisconnect;
      if (options.guard.isCurrent()) options.onFinish(message, needsRecovery);
    },
  });

  const reload = async () => {
    options.guard.assertCurrent();
    const request = new AbortController();
    historyRequest = request;
    const abort = () => request.abort();
    lifetime.signal.addEventListener("abort", abort, { once: true });
    const messages = await options
      .getHistory(request.signal)
      .catch((error: unknown) => {
        if (
          !persisted &&
          error &&
          typeof error === "object" &&
          "status" in error &&
          error.status === 404
        )
          return [];
        throw error;
      })
      .finally(() => {
        lifetime.signal.removeEventListener("abort", abort);
        if (historyRequest === request) historyRequest = null;
      });
    options.guard.assertCurrent();
    if (messages.length) persisted = true;
    chat.messages = messages;
    chat.clearError();
    needsRecovery = false;
  };
  const run = (
    operation: (assertNotCancelled: () => void) => Promise<void>,
  ) => {
    options.guard.assertCurrent();
    if (pending || chat.status === "submitted" || chat.status === "streaming") {
      throw new Error("Wait for Maya to finish or stop the current response.");
    }
    // Lock before the first await, including credential and history lookup.
    const version = cancellationVersion;
    const assertNotCancelled = () => {
      options.guard.assertCurrent();
      if (version !== cancellationVersion)
        throw new Error("Maya request cancelled.");
    };
    const result = Promise.resolve().then(() => {
      assertNotCancelled();
      return operation(assertNotCancelled);
    });
    pending = result;
    return result.finally(() => {
      if (pending === result) pending = null;
    });
  };
  const recover = async () => {
    if (needsRecovery) await reload();
    options.guard.assertCurrent();
  };
  const checkResponse = () => {
    options.guard.assertCurrent();
    if (chat.error) {
      needsRecovery = true;
      throw chat.error;
    }
  };
  const resolveBody = async (body: RequestBody) => {
    if (typeof body !== "function") return body;
    const request = new AbortController();
    contextRequest = request;
    const abort = () => request.abort();
    lifetime.signal.addEventListener("abort", abort, { once: true });
    return body(request.signal).finally(() => {
      lifetime.signal.removeEventListener("abort", abort);
      if (contextRequest === request) contextRequest = null;
    });
  };

  const dispose = () => {
    options.guard.dispose();
    lifetime.abort();
    options.cloud.dispose();
    void chat.stop();
    chat.messages = [];
  };
  return {
    chat,
    guard: options.guard,
    get isBusy() {
      return pending !== null;
    },
    retain: () => {
      const version = ++attachmentVersion;
      return () =>
        queueMicrotask(() => {
          if (version === attachmentVersion) dispose();
        });
    },
    send: (text: string, body: RequestBody, files: FileUIPart[] = []) =>
      run(async (assertNotCancelled) => {
        const content = text.trim();
        const attachments = files.map((file) => ({ ...file }));
        validateMayaFileParts(attachments);
        if (!content && attachments.length === 0)
          throw new Error(
            "Write a message or attach an image or PDF for Maya.",
          );
        if (content.length > 16_000)
          throw new Error("Keep your message under 16,000 characters.");
        await recover();
        assertNotCancelled();
        if (hasUnresolvedMayaApprovals(chat.messages)) {
          throw new Error(
            "Confirm or cancel the proposed change before sending another message.",
          );
        }
        const requestBody = await resolveBody(body);
        assertNotCancelled();
        chat.clearError();
        await chat.sendMessage(
          { ...(content ? { text: content } : {}), files: attachments },
          { body: requestBody },
        );
        checkResponse();
      }),
    approve: (id: string, approved: boolean, body: RequestBody) =>
      run(async (assertNotCancelled) => {
        await recover();
        assertNotCancelled();
        if (
          !getPendingMayaApprovals(chat.messages).some(
            (approval) => approval.id === id,
          )
        ) {
          throw new Error(
            "This confirmation is no longer pending. Review the latest conversation.",
          );
        }
        const requestBody = await resolveBody(body);
        assertNotCancelled();
        await chat.addToolApprovalResponse({ id, approved });
        assertNotCancelled();
        // SDK auto-continuation is deliberately disabled. Only this user action
        // can submit the approved call; mounting/loading/reconnecting never does.
        if (
          lastAssistantMessageIsCompleteWithApprovalResponses({
            messages: chat.messages,
          })
        ) {
          await chat.sendMessage(undefined, { body: requestBody });
          checkResponse();
        }
      }),
    reload: () => run(reload),
    persistVoiceMessages: (
      messages: Pick<VoiceHistoryMessage, "id" | "role" | "text">[],
    ) =>
      run(async (assertNotCancelled) => {
        await recover();
        assertNotCancelled();
        if (hasUnresolvedMayaApprovals(chat.messages))
          throw new Error(
            "Confirm or cancel Maya's proposed change before saving voice messages.",
          );
        const response = await options.cloud.fetch(
          options.cloud.url("/api/chat/voice-history"),
          {
            method: "POST",
            body: JSON.stringify({
              id: options.id,
              workspace: { slug: options.workspace },
              messages,
            }),
          },
        );
        const body = (await response.json()) as { messages?: unknown };
        const history = await safeValidateUIMessages({
          messages: body.messages,
        });
        if (!history.success)
          throw new Error("Maya returned invalid saved voice history.");
        assertNotCancelled();
        chat.messages = history.data;
        persisted = true;
        chat.clearError();
      }),
    stop: async () => {
      cancellationVersion++;
      if (chat.status === "submitted" || chat.status === "streaming")
        needsRecovery = true;
      historyRequest?.abort();
      contextRequest?.abort();
      await chat.stop();
      options.cloud.abort();
      // AI SDK stop only signals abort. Await settlement before replacing history.
      await pending?.catch(() => undefined);
    },
    dispose,
  };
};
