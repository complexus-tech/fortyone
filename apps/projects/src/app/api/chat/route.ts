/* eslint-disable turbo/no-undeclared-env-vars -- ok */
import type { OpenAIResponsesProviderOptions } from "@ai-sdk/openai";
import { createOpenAI } from "@ai-sdk/openai";
import { createGoogleGenerativeAI } from "@ai-sdk/google";
import { devToolsMiddleware } from "@ai-sdk/devtools";
import {
  consumeStream,
  convertToModelMessages,
  generateId,
  stepCountIs,
  streamText,
  wrapLanguageModel,
} from "ai";
import type { NextRequest } from "next/server";
import { withTracing } from "@posthog/ai";
import {
  OPENAI_DEFAULT_REASONING_EFFORT,
  OPENAI_TEXT_MODEL,
} from "@/lib/ai/models";
import { tools } from "@/lib/ai/tools";
import { withCompactModelOutputs } from "@/lib/ai/model-tools";
import { getMayaPromptCacheKey } from "@/lib/ai/prompt-cache-key";
import {
  omitMayaToolSearchHistory,
  withOpenAIToolDiscovery,
} from "@/lib/ai/tool-discovery";
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";
import { systemPrompt } from "./system";
import { getUserContext } from "./user-context";
import { beginChatWrite, saveChat } from "./save-chat";
import { normalizeInlineFileData } from "./normalize-file-data";
import { resolveJoinedTeams } from "./resolve-joined-teams";
import { selectActiveToolPlan } from "./active-tools";
import {
  assertLatestUserTextWithinContextBudget,
  compactChatToolOutputs,
  compactUnknownChatToolOutputs,
  getChatContextStartIndex,
  omitHistoricalChatAttachments,
  pruneChatModelMessages,
} from "./chat-context";
import {
  formatChatErrorDiagnostic,
  getChatErrorDiagnostic,
  getChatStreamErrorMessage,
} from "./chat-errors";
import { hasTerminalMutationResult } from "./stop-conditions";
import { createMutationToolApprovalResponse } from "./mutation-tool-approval";
import { resolveStoryCreationDefaults } from "./story-creation-defaults";
import { sanitizeOpenAIHistoryItemReferences } from "./openai-history";
import {
  runWithMayaHttpRequestContext,
  withMayaHttpRequestContext,
} from "./maya-http-request-context";
import {
  type ChatRequestBody,
  dispatchValidatedChatRequest,
} from "./chat-request";

export const maxDuration = 300;

const CHAT_TIMEOUT = {
  // Tool execution can legitimately be silent for up to 60 seconds. A chunk
  // or step watchdog would abort those healthy calls before their own timeout;
  // the total budget still bounds the complete model-and-tool run and leaves
  // enough function time for preflight reads and an idempotent transcript-
  // finalization retry. Increasing the function ceiling does not delay normal
  // responses; it only prevents valid long-running tool work being killed.
  totalMs: 250_000,
} as const;
const MAX_TOOL_STEPS = 12;
const modelTools = withMayaHttpRequestContext(withCompactModelOutputs(tools));

const getChatResponseErrorMessage = (error: unknown) => {
  const message = getChatStreamErrorMessage(error);
  if (process.env.NODE_ENV !== "development") return message;

  return `${message}\n\nDevelopment diagnostic: ${formatChatErrorDiagnostic(error)}`;
};

const handleChatRequest = async (
  req: NextRequest,
  requestBody: ChatRequestBody,
) => {
  const sessionPromise = auth();
  const {
    messages: messagesFromRequest,
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    id,
    username,
    terminology,
    workspace,
    memories,
    messageId,
    provider = "openai",
    totalMessages,
    trigger,
  } = requestBody;
  const session = await sessionPromise;
  if (!session?.user) {
    return new Response("Unauthorized", { status: 401 });
  }

  const uiMessages = messagesFromRequest;
  const mutationApprovalResponse = createMutationToolApprovalResponse({
    abortSignal: req.signal,
    chatId: id,
    messageId,
    messages: uiMessages,
    userId: session.user.id,
    workspaceSlug: workspace.slug,
  });
  if (mutationApprovalResponse) return mutationApprovalResponse;

  if (messageId && trigger !== "regenerate-message") {
    return new Response(
      "Editing an earlier Maya message is not supported yet. Start a new message or regenerate an assistant response.",
      { status: 409 },
    );
  }

  assertLatestUserTextWithinContextBudget(uiMessages);

  const writeReservation = await beginChatWrite({
    id,
    messageId,
    messages: uiMessages,
    operation: trigger === "regenerate-message" ? "regenerate" : "append",
    workspaceSlug: workspace.slug,
  });
  const canonicalUiMessages = writeReservation.messages ?? uiMessages;

  const messagesWithoutHistoricalAttachments =
    omitHistoricalChatAttachments(canonicalUiMessages);
  const messagesWithoutToolSearchHistory = omitMayaToolSearchHistory(
    messagesWithoutHistoricalAttachments,
  );
  const compactMessages = compactChatToolOutputs(
    messagesWithoutToolSearchHistory,
  );
  const contextStartIndex = getChatContextStartIndex(compactMessages);
  const contextMessages = compactMessages.slice(contextStartIndex);
  const activeToolPlan = selectActiveToolPlan({
    currentPath,
    messages: contextMessages,
    storyTerminology: terminology.stories,
  });
  const openaiClient = createOpenAI({
    apiKey: process.env.OPENAI_API_KEY,
  });
  const googleClient = createGoogleGenerativeAI({
    apiKey: process.env.GOOGLE_API_KEY,
  });
  const runtimeTools =
    provider === "openai"
      ? withOpenAIToolDiscovery(
          modelTools,
          openaiClient.tools.toolSearch({ execution: "server" }),
        )
      : modelTools;
  const runtimeToolNames = new Set(Object.keys(runtimeTools));
  // Compact copies determine the byte-bounded suffix and tool routing. Convert
  // the aligned raw suffix so each registered toModelOutput projector runs
  // exactly once; double-projecting would corrupt stateful tool receipts.
  const convertedMessages = await convertToModelMessages(
    compactUnknownChatToolOutputs(
      messagesWithoutToolSearchHistory.slice(contextStartIndex),
      runtimeToolNames,
    ),
    { tools: runtimeTools },
  );

  const modelMessages = normalizeInlineFileData(
    sanitizeOpenAIHistoryItemReferences(
      pruneChatModelMessages(convertedMessages),
    ),
  );
  const [joinedTeams, storyCreationDefaults] =
    await runWithMayaHttpRequestContext(req.signal, () =>
      Promise.all([
        resolveJoinedTeams({
          session,
          workspaceSlug: workspace.slug,
        }),
        resolveStoryCreationDefaults({
          ctx: { session, workspaceSlug: workspace.slug },
        }),
      ]),
    );

  // Get user context for "me" resolution
  const userContext = getUserContext({
    user: session.user,
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    memories,
    joinedTeams,
    storyCreationDefaults,
    username: username ?? subscription?.username,
    terminology,
    workspace,
    totalMessages,
  });

  const phClient = posthogServer();

  let client =
    provider === "openai"
      ? openaiClient(OPENAI_TEXT_MODEL)
      : googleClient("gemini-3-flash-preview");

  if (process.env.NODE_ENV === "development") {
    client = wrapLanguageModel({
      model: client,
      middleware: devToolsMiddleware(),
    });
  }

  const model = withTracing(client, phClient, {
    posthogDistinctId: session.user.id,
    posthogPrivacyMode: true,
    posthogProperties: {
      active_tool_count: runtimeToolNames.size,
      chat_context_message_count: contextMessages.length,
      conversation_id: id,
      paid: subscription?.status === "active",
      tool_selection_source: activeToolPlan.source,
    },
  });

  const result = streamText({
    abortSignal: req.signal,
    model,
    messages: modelMessages,
    stopWhen: [hasTerminalMutationResult, stepCountIs(MAX_TOOL_STEPS)],
    tools: runtimeTools,
    system: systemPrompt + userContext,
    experimental_context: {
      chatId: id,
      workspaceSlug: workspace.slug,
    },
    providerOptions: {
      openai: {
        promptCacheKey: getMayaPromptCacheKey(workspace.id),
        reasoningEffort: OPENAI_DEFAULT_REASONING_EFFORT,
        textVerbosity: "low",
      } satisfies OpenAIResponsesProviderOptions,
      google: {
        thinkingConfig: {
          thinkingBudget: -1,
          includeThoughts: false,
        },
      },
    },
    onError: ({ error }) => {
      // eslint-disable-next-line no-console -- Diagnostics intentionally omit provider payloads and user content.
      console.error(
        "[chat/route] Stream error",
        formatChatErrorDiagnostic(error),
      );
    },
    onStepFinish: ({ finishReason, toolCalls, usage }) => {
      if (finishReason !== "length" && finishReason !== "error") return;

      // eslint-disable-next-line no-console -- Payload-free diagnostics for provider truncation and failures.
      console.warn("[chat/route] Abnormal model step finish", {
        finishReason,
        inputTokens: usage.inputTokens,
        outputTokens: usage.outputTokens,
        toolCallCount: toolCalls.length,
      });
    },
    timeout: CHAT_TIMEOUT,
  });
  return result.toUIMessageStreamResponse({
    // Drain an independent server-side copy so AI SDK still reaches onFinish
    // and persists the canonical partial response when the browser disconnects
    // or the user deliberately stops a stream.
    consumeSseStream: consumeStream,
    // Persistence requires stable identities for every newly generated
    // assistant message. AI SDK only adds the response ID when this generator
    // is provided; the same ID is emitted to the browser and passed to
    // onFinish for the durable transcript write.
    generateMessageId: generateId,
    sendReasoning: false,
    sendSources: false,
    originalMessages: canonicalUiMessages,
    onFinish: async ({ messages }) => {
      await saveChat({
        id,
        messages,
        reservation: writeReservation,
        workspaceSlug: workspace.slug,
      });
    },
    onError: getChatResponseErrorMessage,
  });
};

export async function POST(req: NextRequest) {
  try {
    return await dispatchValidatedChatRequest({
      handle: (requestBody) => handleChatRequest(req, requestBody),
      request: req,
    });
  } catch (error) {
    // eslint-disable-next-line no-console -- Diagnostics intentionally omit request payloads and user content.
    console.error(
      "[chat/route] Request setup failed",
      JSON.stringify({
        error: getChatErrorDiagnostic(error),
        requestOrigin: req.headers.get("origin") ?? "missing",
      }),
    );
    return new Response(getChatResponseErrorMessage(error), { status: 500 });
  }
}
