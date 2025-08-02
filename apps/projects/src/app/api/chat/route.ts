import { createOpenAI } from "@ai-sdk/openai";
import type { Message } from "ai";
import { appendResponseMessages, generateObject, streamText } from "ai";
import type { NextRequest } from "next/server";
import { z } from "zod";
import { withTracing } from "@posthog/ai";
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
  keyResultsListTool,
  keyResultsCreateTool,
  keyResultsUpdateTool,
  keyResultsDeleteTool,
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
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";
import { systemPrompt } from "./system-xml";
import { getUserContext } from "./user-context";

export const maxDuration = 30;

const saveChat = async ({
  id,
  messages,
}: {
  id: string;
  messages: Message[];
}) => {
  const session = await auth();
  let title = "";
  // if its a new chat generate the title
  const phClient = posthogServer();

  const openaiClient = createOpenAI({
    // eslint-disable-next-line turbo/no-undeclared-env-vars -- this is ok
    apiKey: process.env.OPENAI_API_KEY,
    compatibility: "strict",
  });

  const model = withTracing(openaiClient("gpt-4.1-nano"), phClient, {
    posthogDistinctId: session?.user?.email ?? undefined,
    posthogProperties: {
      conversation_id: id,
    },
  });
  if (messages.length <= 3) {
    const result = await generateObject({
      model,
      schema: z.object({
        title: z.string(),
      }),
      temperature: 0.6,
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
    terminology,
    workspace,
  } = await req.json();

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
    // eslint-disable-next-line turbo/no-undeclared-env-vars -- this is ok
    apiKey: process.env.OPENAI_API_KEY,
    compatibility: "strict",
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
      messages,
      maxSteps: 10,
      maxTokens: 4000,
      temperature: 0.5,
      maxRetries: 2,
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
        keyResultsList: keyResultsListTool,
        keyResultsCreate: keyResultsCreateTool,
        keyResultsUpdate: keyResultsUpdateTool,
        keyResultsDelete: keyResultsDeleteTool,
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
