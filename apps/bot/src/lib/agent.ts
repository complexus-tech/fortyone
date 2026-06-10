import { createOpenAI } from "@ai-sdk/openai";
import { streamText } from "ai";
import { tools } from "@/lib/tools";

const createMissingKeyStream = () =>
  (async function* fallback() {
    yield "OPENAI_API_KEY is missing. Add it to apps/bot/.env.";
  })();

export const createAgentStream = (prompt: string) => {
  const apiKey = process.env.OPENAI_API_KEY;
  const modelName = process.env.BOT_OPENAI_MODEL ?? "gpt-5.4-mini";

  if (!apiKey) {
    return createMissingKeyStream();
  }

  const openai = createOpenAI({ apiKey });
  const result = streamText({
    model: openai(modelName),
    system:
      "You are Photon, the FortyOne Slack assistant. Keep answers concise and helpful. Use tools when relevant.",
    prompt: prompt.trim() || "Help me with my workspace.",
    tools,
    maxOutputTokens: 600,
  });

  return result.fullStream;
};
