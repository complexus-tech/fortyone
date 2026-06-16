import { tool } from "ai";
import { z } from "zod";
import type { SlackActor, StoryRuntime } from "@/lib/runtime";

export const createTools = (runtime: StoryRuntime, actor: SlackActor) => ({
  listStoryFormOptions: tool({
    description:
      "List the teams, statuses, members, objectives, and labels currently available in the Slack create-story form.",
    inputSchema: z.object({}),
    execute: async () => ({
      actor,
      options: await runtime.listStoryOptions(actor),
    }),
  }),
});
