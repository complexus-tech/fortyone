import { createOpenAI } from "@ai-sdk/openai";
import { generateText, streamText } from "ai";
import type { BotConfig } from "@/lib/config";
import type { SlackActor, StoryRuntime } from "@/lib/runtime";
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
  runtime: StoryRuntime,
  actor: SlackActor,
) => {
  return {
    system:
      "You are Maya, the FortyOne Slack assistant. Keep answers concise and useful for Slack. Use tools when relevant. Story creation happens through the Slack create-story form, not free-form chat.",
    prompt: prompt.trim() || "Help me with my workspace.",
    tools: createTools(runtime, actor),
    maxOutputTokens: 600,
  };
};

export const createAgentStream = (
  prompt: string,
  config: BotConfig,
  runtime: StoryRuntime,
  actor: SlackActor,
) => {
  const openai = getOpenAIClient(config);

  if (!openai) {
    return createMissingKeyStream();
  }

  const result = streamText({
    ...createAgentRequest(prompt, config, runtime, actor),
    model: openai(config.openAIModel),
  });

  return result.fullStream;
};

export const generateAgentReply = async (
  prompt: string,
  config: BotConfig,
  runtime: StoryRuntime,
  actor: SlackActor,
) => {
  const openai = getOpenAIClient(config);

  if (!openai) {
    return createMissingKeyText();
  }

  const result = await generateText({
    ...createAgentRequest(prompt, config, runtime, actor),
    model: openai(config.openAIModel),
  });

  return result.text.trim() || "I couldn't generate a response.";
};
