import type { ToolExecutionOptions } from "ai";
import { createUIMessageStream, createUIMessageStreamResponse } from "ai";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import {
  bulkCreateStories,
  bulkCreateStoriesInputSchema,
} from "@/lib/ai/tools/stories/bulk-create-stories";
import {
  createStory,
  createStoryInputSchema,
} from "@/lib/ai/tools/stories/create-story";
import { saveChat } from "./save-chat";

const STORY_TOOL_TYPES = new Set([
  "tool-createStory",
  "tool-bulkCreateStories",
]);

type StoryToolName = "bulkCreateStories" | "createStory";

type StoryToolApproval = {
  approved: boolean;
  input: unknown;
  toolCallId: string;
  toolName: StoryToolName;
};

const getStoryToolApprovals = (
  messages: MayaUIMessage[],
): StoryToolApproval[] => {
  const lastMessage = messages.at(-1);
  if (lastMessage?.role !== "assistant") return [];

  return lastMessage.parts.flatMap((part) => {
    if (!STORY_TOOL_TYPES.has(part.type) || !("state" in part)) return [];
    if (part.state !== "approval-responded" || !("approval" in part)) {
      return [];
    }

    const approval = part.approval;
    if (
      typeof approval.approved !== "boolean" ||
      !("input" in part) ||
      !("toolCallId" in part) ||
      typeof part.toolCallId !== "string"
    ) {
      return [];
    }

    return [
      {
        approved: approval.approved,
        input: part.input,
        toolCallId: part.toolCallId,
        toolName:
          part.type === "tool-createStory"
            ? ("createStory" as const)
            : ("bulkCreateStories" as const),
      },
    ];
  });
};

const executeApprovedStoryTool = async ({
  approval,
  abortSignal,
  workspaceSlug,
}: {
  approval: StoryToolApproval;
  abortSignal: AbortSignal;
  workspaceSlug: string;
}) => {
  const options: ToolExecutionOptions = {
    abortSignal,
    experimental_context: { workspaceSlug },
    messages: [],
    toolCallId: approval.toolCallId,
  };

  if (approval.toolName === "createStory") {
    const input = createStoryInputSchema.parse(approval.input);
    if (!createStory.execute) throw new Error("Story creation is unavailable.");
    return createStory.execute(input, options);
  }

  const input = bulkCreateStoriesInputSchema.parse(approval.input);
  if (!bulkCreateStories.execute) {
    throw new Error("Bulk story creation is unavailable.");
  }
  return bulkCreateStories.execute(input, options);
};

const getApprovalFailureOutput = () => ({
  error:
    "The story details could not be validated. Ask Maya to prepare the stories again.",
  success: false,
});

export const createStoryToolApprovalResponse = ({
  abortSignal,
  chatId,
  messages,
  workspaceSlug,
}: {
  abortSignal: AbortSignal;
  chatId: string;
  messages: MayaUIMessage[];
  workspaceSlug: string;
}): Response | undefined => {
  const approvals = getStoryToolApprovals(messages);
  if (approvals.length === 0) return undefined;

  const stream = createUIMessageStream<MayaUIMessage>({
    execute: async ({ writer }) => {
      const results = await Promise.all(
        approvals.map(async (approval) => {
          if (!approval.approved) {
            return { approval, denied: true as const };
          }

          try {
            return {
              approval,
              denied: false as const,
              output: await executeApprovedStoryTool({
                abortSignal,
                approval,
                workspaceSlug,
              }),
            };
          } catch (error) {
            // eslint-disable-next-line no-console -- Keep the validation failure trace server-side.
            console.error("[chat/route] Story approval failed:", error);
            return {
              approval,
              denied: false as const,
              output: getApprovalFailureOutput(),
            };
          }
        }),
      );

      for (const result of results) {
        if (result.denied) {
          writer.write({
            toolCallId: result.approval.toolCallId,
            type: "tool-output-denied",
          });
          continue;
        }

        writer.write({
          output: result.output,
          toolCallId: result.approval.toolCallId,
          type: "tool-output-available",
        });
      }
    },
    onFinish: async ({ messages: finishedMessages }) => {
      await saveChat({
        generateTitle: false,
        id: chatId,
        messages: finishedMessages,
        workspaceSlug,
      });
    },
    originalMessages: messages,
  });

  return createUIMessageStreamResponse({ stream });
};
