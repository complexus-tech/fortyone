import { openai } from "@ai-sdk/openai";
import type { Message } from "ai";
import { appendResponseMessages, generateObject, streamText } from "ai";
import type { NextRequest } from "next/server";
import { z } from "zod";
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

const saveChat = async ({
  id,
  messages,
}: {
  id: string;
  messages: Message[];
}) => {
  let title = "";
  // if its a new chat generate the title
  if (messages.length <= 3) {
    const result = await generateObject({
      model: openai("gpt-4.1-nano"),
      schema: z.object({
        title: z.string(),
      }),
      prompt: `You're generating a short title for a conversation in Complexus, a project management platform. Use the first user message to infer what the chat is about. Keep the title short, clear, and relevant to project work (e.g. planning, tasks, bugs, OKRs).
    
    User message:
    "${messages[0].content}"

    Title:`,
    });
    title = result.object.title;
  }

  console.log("Saving chat", id, messages, title);
};

export async function POST(req: NextRequest) {
  const {
    messages,
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    teams,
    id,
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

    async onFinish({ response }) {
      await saveChat({
        id,
        messages: appendResponseMessages({
          messages,
          responseMessages: response.messages,
        }),
      });
    },
  });
  return result.toDataStreamResponse({
    getErrorMessage: () => {
      return "I'm having trouble connecting to my AI service right now. You can ask me to help you navigate the app, manage stories, get sprint insights, and provide team information.";
    },
  });
}
