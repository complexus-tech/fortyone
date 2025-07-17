import { openai } from "@ai-sdk/openai";
import { streamText } from "ai";
import type { NextRequest } from "next/server";
import {
  navigation,
  theme,
  quickCreate,
  teamsTool,
  membersTool,
  storiesTool,
  statusesTool,
  sprintsTool,
  objectivesTool,
  objectiveStatusesTool,
  searchTool,
  notificationsTool,
} from "@/lib/ai/tools";
import { systemPrompt } from "./system";
import { getUserContext } from "./user-context";

export const maxDuration = 30;

export async function POST(req: NextRequest) {
  const {
    messages,
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    teams,
  } = await req.json();

  // Get user context for "me" resolution
  const userContext = await getUserContext({
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    teams,
  });

  const result = streamText({
    model: openai("gpt-4.1-mini"),
    messages,
    maxSteps: 10,
    maxTokens: 4000,
    tools: {
      navigation,
      theme,
      quickCreate,
      teams: teamsTool,
      members: membersTool,
      stories: storiesTool,
      statuses: statusesTool,
      sprints: sprintsTool,
      objectives: objectivesTool,
      objectiveStatuses: objectiveStatusesTool,
      search: searchTool,
      notifications: notificationsTool,
    },
    system: systemPrompt + userContext,
  });
  return result.toDataStreamResponse({
    getErrorMessage: () => {
      return "I'm having trouble connecting to my AI service right now. You can ask me to help you navigate the app, manage stories, get sprint insights, and provide team information.";
    },
  });
}
