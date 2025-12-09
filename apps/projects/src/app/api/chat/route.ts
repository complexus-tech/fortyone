/* eslint-disable turbo/no-undeclared-env-vars -- ok */
import type { OpenAIResponsesProviderOptions } from "@ai-sdk/openai";
import { createOpenAI } from "@ai-sdk/openai";
import { createGoogleGenerativeAI, google } from "@ai-sdk/google";
import type { Tool, UIMessage } from "ai";
import {
  convertToModelMessages,
  stepCountIs,
  streamText,
  smoothStream,
} from "ai";
import type { NextRequest } from "next/server";
import { withTracing } from "@posthog/ai";
import { tools } from "@/lib/ai/tools";
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";
import { systemPrompt } from "./system";
import { getUserContext } from "./user-context";
import { saveChat } from "./save-chat";

export const maxDuration = 30;

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
    webSearchEnabled = false,
    provider = "openai",
  } = await req.json();
  const modelMessages = convertToModelMessages(
    messagesFromRequest as UIMessage[],
  );

  // Get user context for "me" resolution
  const userContext = await getUserContext({
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    teams,
    username,
    terminology,
    workspace,
  });

  const session = await auth();

  const phClient = posthogServer();

  const openaiClient = createOpenAI({
    apiKey: process.env.OPENAI_API_KEY,
  });
  const googleClient = createGoogleGenerativeAI({
    apiKey: process.env.GOOGLE_API_KEY,
  });

  const client =
    provider === "openai"
      ? openaiClient("gpt-4.1-mini")
      : googleClient("gemini-2.5-pro");

  const model = withTracing(client, phClient, {
    posthogDistinctId: session?.user?.email ?? undefined,
    posthogProperties: {
      conversation_id: id,
      paid: subscription?.status === "active",
    },
  });

  try {
    const result = streamText({
      model,
      messages: modelMessages,
      maxOutputTokens: 4000,
      stopWhen: [stepCountIs(15)],
      tools: {
        ...tools,
        ...(webSearchEnabled
          ? { google_search: google.tools.googleSearch({}) as Tool }
          : {}),
      },
      system: systemPrompt + userContext,
      experimental_transform: smoothStream({
        delayInMs: 20,
        chunking: "word",
      }),
      providerOptions: {
        // openai: {
        //   reasoningEffort: "low",
        //   reasoningSummary: "auto",
        //   textVerbosity: "low",
        // } satisfies OpenAIResponsesProviderOptions,
        google: {
          thinkingConfig: {
            thinkingBudget: -1,
            includeThoughts: true,
          },
        },
      },
    });
    return result.toUIMessageStreamResponse({
      sendReasoning: true,
      sendSources: true,
      originalMessages: messagesFromRequest,
      onFinish: ({ messages }) => {
        saveChat({ id, messages });
      },
    });
  } catch {
    throw new Error(
      "I'm having trouble connecting to my AI service right now. You can ask me to help you navigate the app, manage stories, get sprint insights, and provide team information.",
    );
  }
}
