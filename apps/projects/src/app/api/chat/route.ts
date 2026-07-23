/* eslint-disable turbo/no-undeclared-env-vars -- ok */
import type { OpenAIResponsesProviderOptions } from "@ai-sdk/openai";
import { createOpenAI } from "@ai-sdk/openai";
import { createGoogleGenerativeAI } from "@ai-sdk/google";
import { devToolsMiddleware } from "@ai-sdk/devtools";
import type { UIMessage } from "ai";
import {
  convertToModelMessages,
  stepCountIs,
  streamText,
  wrapLanguageModel,
} from "ai";
import type { NextRequest } from "next/server";
import { withTracing } from "@posthog/ai";
import { tools } from "@/lib/ai/tools";
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";
import { systemPrompt } from "./system";
import { getUserContext } from "./user-context";
import { saveChat } from "./save-chat";
import { normalizeInlineFileData } from "./normalize-file-data";

export const maxDuration = 120;

const MAX_OUTPUT_TOKENS = 2000;
const MAX_TOOL_STEPS = 8;
const MAYA_PROMPT_CACHE_NAMESPACE = "maya-projects-v1";

export async function POST(req: NextRequest) {
  const {
    messages: messagesFromRequest,
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    teams,
    id,
    username,
    terminology,
    workspace,
    memories,
    provider = "openai",
    totalMessages,
  } = await req.json();

  const [convertedMessages, session] = await Promise.all([
    convertToModelMessages(messagesFromRequest as UIMessage[]),
    auth(),
  ]);
  const modelMessages = normalizeInlineFileData(convertedMessages);

  // Get user context for "me" resolution
  const userContext = getUserContext({
    user: session?.user,
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    memories,
    teams,
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
      ? openaiClient("gpt-5.6-luna")
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
      conversation_id: id,
      paid: subscription?.status === "active",
    },
  });

  try {
    const result = streamText({
      model,
      messages: modelMessages,
      maxOutputTokens: MAX_OUTPUT_TOKENS,
      stopWhen: [stepCountIs(MAX_TOOL_STEPS)],
      tools: {
        ...tools,
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
          reasoningEffort: "low",
          textVerbosity: "low",
        } satisfies OpenAIResponsesProviderOptions,
        google: {
          thinkingConfig: {
            thinkingBudget: -1,
            includeThoughts: false,
          },
        },
      },
    });
    return result.toUIMessageStreamResponse({
      sendReasoning: false,
      sendSources: false,
      originalMessages: messagesFromRequest,
      onFinish: async ({ messages }) => {
        await saveChat({ id, messages, workspaceSlug: workspace?.slug || "" });
      },
    });
  } catch (error) {
    // eslint-disable-next-line no-console -- Preserve server-side diagnostics.
    console.error("[chat/route] Stream error:", error);
    throw new Error(
      "I'm having trouble connecting to my AI service right now. You can ask me to help you navigate the app, manage stories, get sprint insights, and provide team information.",
    );
  }
}
