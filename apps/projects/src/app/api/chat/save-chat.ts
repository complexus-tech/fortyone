import type { OpenAIResponsesProviderOptions } from "@ai-sdk/openai";
import { createOpenAI } from "@ai-sdk/openai";
import type { UIMessage } from "ai";
import { generateText } from "ai";
import { OPENAI_TEXT_MODEL } from "@/lib/ai/models";
import { saveAiChatMessagesAction } from "@/modules/ai-chats/actions/save-ai-chat-messages";
import { createAiChatAction } from "@/modules/ai-chats/actions/create-ai-chat";
import {
  getChatTitle,
  getChatTitleSource,
  normalizeGeneratedChatTitle,
} from "./chat-title";

const TITLE_MAX_OUTPUT_TOKENS = 32;

const generateChatTitle = async (messages: UIMessage[]) => {
  const fallbackTitle = getChatTitle(messages);
  const titleSource = getChatTitleSource(messages);
  if (!titleSource) return fallbackTitle;

  try {
    const openai = createOpenAI({
      // eslint-disable-next-line turbo/no-undeclared-env-vars -- OpenAI is the configured chat provider.
      apiKey: process.env.OPENAI_API_KEY,
    });
    const result = await generateText({
      model: openai(OPENAI_TEXT_MODEL),
      maxRetries: 1,
      maxOutputTokens: TITLE_MAX_OUTPUT_TOKENS,
      prompt: `Write a clear project-management chat title in at most 8 words. Return only the title, without quotes.\n\nUser request: ${titleSource}`,
      providerOptions: {
        openai: {
          reasoningEffort: "low",
          textVerbosity: "low",
        } satisfies OpenAIResponsesProviderOptions,
      },
    });

    return normalizeGeneratedChatTitle(result.text) || fallbackTitle;
  } catch {
    return fallbackTitle;
  }
};

export const saveChat = async ({
  id,
  messages,
  workspaceSlug,
}: {
  id: string;
  messages: UIMessage[];
  workspaceSlug: string;
}) => {
  const title = messages.length <= 3 ? await generateChatTitle(messages) : "";

  try {
    if (title) {
      await createAiChatAction({ id, title, messages }, workspaceSlug);
    } else {
      await saveAiChatMessagesAction({ id, messages }, workspaceSlug);
    }
  } catch (error) {
    // log to analytics later
  }
};
