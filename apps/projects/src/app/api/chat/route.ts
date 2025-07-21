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
  commentsTool,
  attachmentsTool,
  storyActivitiesTool,
  linksTool,
  labelsTool,
  storyLabelsTool,
} from "@/lib/ai/tools";
import { saveAiChatMessagesAction } from "@/modules/ai-chats/actions/save-ai-chat-messages";
import { createAiChatAction } from "@/modules/ai-chats/actions/create-ai-chat";
import { suggestions } from "@/lib/ai/tools/suggestions";
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

  try {
    if (title) {
      await createAiChatAction({ id, title, messages });
    } else {
      await saveAiChatMessagesAction({ id, messages });
    }
  } catch (error) {
    // log to posthog or sentry later
  }
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
    username,
  } = await req.json();

  // Get user context for "me" resolution
  const userContext = await getUserContext({
    currentPath,
    currentTheme,
    resolvedTheme,
    subscription,
    teams,
    username,
  });

  try {
    const result = streamText({
      model: openai("gpt-4.1-mini"),
      messages,
      maxSteps: 10,
      maxTokens: 4000,
      tools: {
        navigation,
        theme,
        quickCreate,
        suggestions,
        teams: teamsTool,
        members: membersTool,
        stories: storiesTool,
        statuses: statusesTool,
        sprints: sprintsTool,
        objectives: objectivesTool,
        objectiveStatuses: objectiveStatusesTool,
        search: searchTool,
        notifications: notificationsTool,
        comments: commentsTool,
        attachments: attachmentsTool,
        storyActivities: storyActivitiesTool,
        links: linksTool,
        labels: labelsTool,
        storyLabels: storyLabelsTool,
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
  } catch {
    throw new Error(
      "I'm having trouble connecting to my AI service right now. You can ask me to help you navigate the app, manage stories, get sprint insights, and provide team information.",
    );
  }
}
