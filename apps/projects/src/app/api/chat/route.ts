/* eslint-disable turbo/no-undeclared-env-vars -- this is ok */
import { createOpenAI } from "@ai-sdk/openai";
import type { UIMessage } from "ai";
import {
  convertToModelMessages,
  stepCountIs,
  streamText,
  hasToolCall,
  smoothStream,
} from "ai";
import type { NextRequest } from "next/server";
import { withTracing } from "@posthog/ai";
import { tools } from "@/lib/ai/tools";
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";
import { systemPrompt } from "./system-xml";
import { getUserContext } from "./user-context";
import { saveChat } from "./save-chat";

export const maxDuration = 30;

export async function POST(req: NextRequest) {
  const {
    messages,
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    teams,
    id,
    username,
    terminology,
    workspace,
  } = await req.json();
  const modelMessages = convertToModelMessages(messages as UIMessage[]);

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

  const model = withTracing(openaiClient("gpt-4.1-mini"), phClient, {
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
      temperature: 0.5,
      stopWhen: [stepCountIs(10), hasToolCall("suggestions")],
      tools,
      system: systemPrompt + userContext,
      experimental_transform: smoothStream({
        delayInMs: 20,
        chunking: "word",
      }),
      providerOptions: {
        openai: {
          reasoningEffort: "low",
        },
      },
    });
    return result.toUIMessageStreamResponse({
      originalMessages: messages,
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
