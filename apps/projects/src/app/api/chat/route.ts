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
  membersTool,
  statusesTool,
  objectivesTool,
  objectiveStatusesTool,
  keyResultsListTool,
  keyResultsCreateTool,
  keyResultsUpdateTool,
  keyResultsDeleteTool,
  searchTool,
  notificationsTool,
  commentsTool,
  storyActivitiesTool,
  linksTool,
  labelsTool,
  storyLabelsTool,
} from "@/lib/ai/tools";
import { saveAiChatMessagesAction } from "@/modules/ai-chats/actions/save-ai-chat-messages";
import { createAiChatAction } from "@/modules/ai-chats/actions/create-ai-chat";
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";
import { listAttachments, deleteAttachment } from "@/lib/ai/tools/attachments";
import {
  createObjectiveTool,
  deleteObjectiveTool,
  listObjectivesTool,
  listTeamObjectivesTool,
  updateObjectiveTool,
  getObjectiveDetailsTool,
  objectiveAnalyticsTool,
} from "@/lib/ai/tools/objectives";
import {
  listTeamStories,
  searchStories,
  getStoryDetails,
  createStory,
  updateStory,
  deleteStory,
  bulkUpdateStories,
  bulkDeleteStories,
  bulkCreateStories,
  assignStoriesToUser,
  duplicateStory,
  restoreStory,
  listDueSoon,
  listOverdue,
  listDueToday,
  listDueTomorrow,
} from "@/lib/ai/tools/stories";
import {
  listSprints,
  listRunningSprints,
  getSprintDetailsTool,
  createSprint,
} from "@/lib/ai/tools/sprints";
import {
  listTeams,
  listPublicTeams,
  getTeamDetails,
  listTeamMembers,
  createTeamTool,
  updateTeam,
  joinTeam,
  deleteTeam,
  leaveTeam,
} from "@/lib/ai/tools/teams";
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
        // Teams
        listTeams,
        listPublicTeams,
        getTeamDetails,
        listTeamMembers,
        createTeamTool,
        updateTeam,
        joinTeam,
        deleteTeam,
        leaveTeam,
        members: membersTool,
        // Stories
        listTeamStories,
        searchStories,
        getStoryDetails,
        createStory,
        updateStory,
        deleteStory,
        bulkUpdateStories,
        bulkDeleteStories,
        bulkCreateStories,
        assignStoriesToUser,
        duplicateStory,
        restoreStory,
        listDueSoon,
        listOverdue,
        listDueToday,
        listDueTomorrow,
        statuses: statusesTool,
        // Sprints
        listSprints,
        listRunningSprints,
        getSprintDetailsTool,
        createSprint,
        objectives: objectivesTool,
        objectiveStatuses: objectiveStatusesTool,
        keyResultsListTool,
        keyResultsCreateTool,
        keyResultsUpdateTool,
        keyResultsDeleteTool,
        search: searchTool,
        notifications: notificationsTool,
        comments: commentsTool,
        // Attachments
        listAttachments,
        deleteAttachment,
        // Objectives
        listObjectivesTool,
        listTeamObjectivesTool,
        createObjectiveTool,
        updateObjectiveTool,
        deleteObjectiveTool,
        objectiveAnalyticsTool,
        getObjectiveDetailsTool,
        // Links
        links: linksTool,
        labels: labelsTool,
        storyActivities: storyActivitiesTool,
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
