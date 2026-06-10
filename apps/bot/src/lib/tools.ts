import { tool } from "ai";
import { z } from "zod";

export const tools = {
  getWorkspaceStatus: tool({
    description: "Get a quick workspace status summary.",
    inputSchema: z.object({
      team: z.string().optional().describe("Optional team name or code."),
    }),
    execute: async ({ team }) => {
      return {
        workspace: "FortyOne",
        team: team ?? "General",
        activeStories: 18,
        blockedStories: 2,
        summary: "Sprint is healthy with a small set of blockers to review.",
      };
    },
  }),
};
