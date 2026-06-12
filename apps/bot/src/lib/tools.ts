import { tool } from "ai";
import { z } from "zod";

import type { FortyOneClient, SlackActor } from "@/lib/fortyone-client";

export const createTools = (client: FortyOneClient, actor: SlackActor) => ({
  getMentionNotifications: tool({
    description:
      "Get the current user's recent FortyOne @-mention notifications for Slack.",
    inputSchema: z.object({}),
    execute: async () => client.listMentionNotifications(actor),
  }),
});
