import { createOpenAI } from "@ai-sdk/openai";
import type { UIMessage } from "ai";
import { generateObject } from "ai";
import { z } from "zod";
import { withTracing } from "@posthog/ai";
import { saveAiChatMessagesAction } from "@/modules/ai-chats/actions/save-ai-chat-messages";
import { createAiChatAction } from "@/modules/ai-chats/actions/create-ai-chat";
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";

export const saveChat = async ({
  id,
  messages,
}: {
  id: string;
  messages: UIMessage[];
}) => {
  const session = await auth();
  let title = "";
  // if its a new chat generate the title
  const phClient = posthogServer();

  const openaiClient = createOpenAI({
    // eslint-disable-next-line turbo/no-undeclared-env-vars -- this is ok
    apiKey: process.env.OPENAI_API_KEY,
  });

  const model = withTracing(openaiClient("gpt-4.1-nano"), phClient, {
    posthogDistinctId: session?.user?.email ?? undefined,
    posthogProperties: {
      conversation_id: id,
    },
  });
  if (messages.length <= 3) {
    const firstMessage = messages[0];
    let messageContent = "";
    firstMessage.parts.forEach((part) => {
      if (part.type === "text") {
        messageContent = part.text;
      }
    });

    const result = await generateObject({
      model,
      schema: z.object({
        title: z.string(),
      }),
      temperature: 0.6,
      prompt: `You're generating a short title for a conversation in Complexus, a project management platform. Use the first user message to infer what the chat is about. Keep the title short, clear, and relevant to project work (e.g. planning, tasks, bugs, OKRs).
    
      User message: "${messageContent}"`,
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
