import { z } from "zod";
import { tool } from "ai";

export const navigation = tool({
  description: "Navigate to different pages of the application",
  parameters: z.object({
    destination: z
      .enum([
        "my-work",
        "summary",
        "analytics",
        "sprints",
        "notifications",
        "settings",
        "roadmap",
      ])
      .describe("The destination to navigate to"),
  }),
  execute: async ({ destination }: { destination: string }) => {
    const routes = {
      "my-work": "/my-work",
      summary: "/summary",
      analytics: "/analytics",
      sprints: "/sprints",
      notifications: "/notifications",
      settings: "/settings",
      roadmap: "/roadmap",
    } as const;

    return {
      route: routes[destination as keyof typeof routes],
    };
  },
});
