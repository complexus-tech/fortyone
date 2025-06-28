import { openai } from "@ai-sdk/openai";
import { streamText } from "ai";
import type { NextRequest } from "next/server";
import { navigationTool } from "@/lib/ai/tools";
import { systemPrompt } from "./system";

export async function POST(req: NextRequest) {
  const { messages } = await req.json();

  const result = streamText({
    model: openai("gpt-4o-mini"),
    messages,
    tools: {
      navigate: navigationTool,
    },
    system: systemPrompt,
  });
  return result.toDataStreamResponse({
    getErrorMessage: () => {
      return "I'm having trouble connecting to my AI service right now. You can ask me to help you navigate the app, manage stories, get sprint insights, and provide project management assistance.";
    },
  });
}
