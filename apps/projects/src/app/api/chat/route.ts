/* eslint-disable turbo/no-undeclared-env-vars -- ok */
import type { OpenAIResponsesProviderOptions } from "@ai-sdk/openai";
import { createOpenAI } from "@ai-sdk/openai";
import { createGoogleGenerativeAI } from "@ai-sdk/google";
import { devToolsMiddleware } from "@ai-sdk/devtools";
import {
  convertToModelMessages,
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
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";
import { systemPrompt } from "./system";
import { getUserContext } from "./user-context";
import { saveChat } from "./save-chat";
import { normalizeInlineFileData } from "./normalize-file-data";
import { resolveJoinedTeams } from "./resolve-joined-teams";
import { selectActiveTools } from "./active-tools";
import {
  compactChatToolOutputs,
  pruneChatModelMessages,
  selectRecentChatMessages,
} from "./chat-context";
import { getChatStreamErrorMessage } from "./chat-errors";
import { hasTerminalMutationResult } from "./stop-conditions";
import { createStoryToolApprovalResponse } from "./story-tool-approval";

export const maxDuration = 120;

const MAX_OUTPUT_TOKENS = 3000;
const MAX_TOOL_STEPS = 12;
const MAYA_PROMPT_CACHE_NAMESPACE = "maya-projects-v2";
const modelTools = withCompactModelOutputs(tools);

export async function POST(req: NextRequest) {
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
    provider = "openai",
    totalMessages,
  } = await req.json();

  const uiMessages = messagesFromRequest as MayaUIMessage[];
  const storyApprovalResponse = createStoryToolApprovalResponse({
    abortSignal: req.signal,
    chatId: id,
    messages: uiMessages,
    workspaceSlug: workspace?.slug ?? "",
  });
  if (storyApprovalResponse) return storyApprovalResponse;

  const recentMessages = selectRecentChatMessages(uiMessages);
  const compactMessages = compactChatToolOutputs(recentMessages);
  const activeTools = selectActiveTools({
    currentPath,
    messages: recentMessages,
  });
  const [convertedMessages, session] = await Promise.all([
    convertToModelMessages(compactMessages),
    auth(),
  ]);
  const modelMessages = pruneChatModelMessages(
    normalizeInlineFileData(convertedMessages),
  );
  const joinedTeams = await resolveJoinedTeams({
    session,
    workspaceSlug: workspace?.slug,
  });

  // Get user context for "me" resolution
  const userContext = getUserContext({
    user: session?.user,
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    memories,
    joinedTeams,
    username: username ?? subscription?.username,
    terminology,
    workspace,
    totalMessages,
  });

  const phClient = posthogServer();

  const openaiClient = createOpenAI({
    apiKey: process.env.OPENAI_API_KEY,
  });
  const googleClient = createGoogleGenerativeAI({
    apiKey: process.env.GOOGLE_API_KEY,
  });

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
    posthogDistinctId: session?.user.email ?? undefined,
    posthogProperties: {
      active_tool_count: activeTools.length,
      chat_context_message_count: recentMessages.length,
      conversation_id: id,
      paid: subscription?.status === "active",
    },
  });

  try {
    const result = streamText({
      model,
      messages: modelMessages,
      maxOutputTokens: MAX_OUTPUT_TOKENS,
      activeTools,
      stopWhen: [hasTerminalMutationResult, stepCountIs(MAX_TOOL_STEPS)],
      tools: {
        ...modelTools,
        // ...(webSearchEnabled
        //   ? {
        //       google_search: google.tools.googleSearch({}) as Tool,
        //     }
        //   : {}),
      },
      system: systemPrompt + userContext,
      experimental_context: {
        workspaceSlug: workspace?.slug,
      },
      providerOptions: {
        openai: {
          promptCacheKey: `${MAYA_PROMPT_CACHE_NAMESPACE}:${workspace?.id ?? "unknown"}`,
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
        // eslint-disable-next-line no-console -- Keep upstream details server-side while the client receives a safe message.
        console.error("[chat/route] Stream error:", error);
      },
    });
    return result.toUIMessageStreamResponse({
      sendReasoning: false,
      sendSources: false,
      originalMessages: messagesFromRequest,
      onFinish: async ({ messages }) => {
        await saveChat({ id, messages, workspaceSlug: workspace?.slug || "" });
      },
      onError: getChatStreamErrorMessage,
    });
  } catch (error) {
    // eslint-disable-next-line no-console -- Preserve server-side diagnostics.
    console.error("[chat/route] Stream error:", error);
    throw new Error(
      "I'm having trouble connecting to my AI service right now. You can ask me to help you navigate the app, manage stories, get sprint insights, and provide team information.",
    );
  }
}
