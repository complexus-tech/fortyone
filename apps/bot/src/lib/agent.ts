import { createOpenAI } from "@ai-sdk/openai";
import { generateText, streamText } from "ai";
import type { BotConfig } from "@/lib/config";
import type { FortyOneClient, SlackActor } from "@/lib/fortyone-client";
import { createTools } from "@/lib/tools";

const createMissingKeyStream = () =>
  (async function* fallback() {
    await Promise.resolve();
    yield "OPENAI_API_KEY is missing. Add it to apps/bot/.env.";
  })();

const createMissingKeyText = () =>
  "OPENAI_API_KEY is missing. Add it to apps/bot/.env.";

const getOpenAIClient = (config: BotConfig) =>
  config.openAIKey ? createOpenAI({ apiKey: config.openAIKey }) : null;

const createAgentRequest = (
  prompt: string,
  config: BotConfig,
  client: FortyOneClient,
  actor: SlackActor,
) => {
  return {
    system:
      "You are Maya, the FortyOne Slack assistant. Keep answers concise and useful for Slack. Use tools when relevant. Do not create or update stories unless a tool explicitly does it.",
    prompt: prompt.trim() || "Help me with my workspace.",
    tools: createTools(client, actor),
    maxOutputTokens: 600,
  };
};

export const createAgentStream = (
  prompt: string,
  config: BotConfig,
  client: FortyOneClient,
  actor: SlackActor,
) => {
  const openai = getOpenAIClient(config);

  if (!openai) {
    return createMissingKeyStream();
  }

  const result = streamText({
    ...createAgentRequest(prompt, config, client, actor),
    model: openai(config.openAIModel),
  });

  return result.fullStream;
};

export const generateAgentReply = async (
  prompt: string,
  config: BotConfig,
  client: FortyOneClient,
  actor: SlackActor,
) => {
  const openai = getOpenAIClient(config);

  if (!openai) {
    return createMissingKeyText();
  }

  const result = await generateText({
    ...createAgentRequest(prompt, config, client, actor),
    model: openai(config.openAIModel),
  });

  return result.text.trim() || "I couldn't generate a response.";
};
