import { z } from "zod";
import { tool } from "ai";

export const navigationTool = tool({
  description: "Navigate to different pages of the application",
  parameters: z.object({
    destination: z
      .enum([
        "my-work",
        "summary",
        "analytics",
        "objectives",
        "sprints",
        "notifications",
        "settings",
        "teams",
        "roadmaps",
      ])
      .describe("The destination to navigate to"),
    reason: z.string().optional().describe("Optional reason for navigation"),
  }),
  execute: async ({
    destination,
    reason,
  }: {
    destination: string;
    reason?: string;
  }) => {
    const routeMap: Record<string, string> = {
      "my-work": "/my-work",
      summary: "/summary",
      analytics: "/analytics",
      objectives: "/objectives",
      sprints: "/sprints",
      notifications: "/notifications",
      settings: "/settings",
      teams: "/teams",
      roadmaps: "/roadmaps",
    };

    const route = routeMap[destination];

    if (!route) {
      return {
        success: false,
        error: `Unknown destination: ${destination}`,
      };
    }

    return {
      success: true,
      route,
      destination,
      reason: reason || `Navigating to ${destination}`,
      message: `I'll take you to the ${destination} page.`,
    };
  },
});
