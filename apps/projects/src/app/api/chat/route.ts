import { openai } from "@ai-sdk/openai";
import { streamText } from "ai";
import type { NextRequest } from "next/server";
import {
  navigation,
  theme,
  quickCreate,
  teamsTool,
  storiesTool,
  statusesTool,
  sprintsTool,
  objectivesTool,
} from "@/lib/ai/tools";
import { systemPrompt } from "./system";

export const maxDuration = 30;

export async function POST(req: NextRequest) {
  const { messages } = await req.json();

  const result = streamText({
    model: openai("gpt-4o-mini"),
    messages,
    maxSteps: 10,
    tools: {
      navigation,
      theme,
      quickCreate,
      teams: teamsTool,
      stories: storiesTool,
      statuses: statusesTool,
      sprints: sprintsTool,
      objectives: objectivesTool,
    },
    system: systemPrompt,
  });
  return result.toDataStreamResponse({
    getErrorMessage: () => {
      return "I'm having trouble connecting to my AI service right now. You can ask me to help you navigate the app, manage stories, get sprint insights, and provide team information.";
    },
  });
}
