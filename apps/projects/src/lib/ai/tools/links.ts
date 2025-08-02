import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getLinks } from "@/lib/queries/links/get-links";
import { createLinkAction } from "@/lib/actions/links/create-link";
import { updateLinkAction } from "@/lib/actions/links/update-link";
import { deleteLinkAction } from "@/lib/actions/links/delete-link";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { getCurrentWorkspace } from "@/lib/hooks/workspaces";

export const linksTool = tool({
  description:
    "Manage story links: list, add, update, delete URLs associated with stories. Supports metadata extraction and link organization.",
  parameters: z.object({
    action: z
      .enum(["list-links", "add-link", "update-link", "delete-link"])
      .describe("The link operation to perform"),

    storyId: z.string().describe("Story ID for link operations"),

    linkId: z
      .string()
      .optional()
      .describe("Link ID for specific link operations"),

    url: z.string().optional().describe("URL for the link"),

    title: z.string().optional().describe("Title for the link"),

    limit: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Limit number of links returned (default: 20, max: 100)"),
  }),

  execute: async ({ action, storyId, linkId, url, title, limit = 20 }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access links",
        };
      }

      // Get user's workspace and role for permissions
      const workspaces = await getWorkspaces(session.token);
      const workspace = getCurrentWorkspace(workspaces);
      const userRole = workspace?.userRole;

      if (!userRole) {
        return {
          success: false,
          error: "Unable to determine user permissions",
        };
      }

      switch (action) {
        case "list-links": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for listing links",
            };
          }

          const links = await getLinks(storyId, session);

          // Limit results
          const limitedLinks = links.slice(0, limit);

          const formattedLinks = limitedLinks.map((link) => ({
            id: link.id,
            url: link.url,
            title: link.title,
            createdAt: link.createdAt,
            updatedAt: link.updatedAt,
          }));

          return {
            success: true,
            links: formattedLinks,
            count: formattedLinks.length,
            totalCount: links.length,
            message: `Found ${formattedLinks.length} link${formattedLinks.length !== 1 ? "s" : ""} for this story.`,
          };
        }

        case "add-link": {
          if (!storyId || !url) {
            return {
              success: false,
              error: "Story ID and URL are required for adding links",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot add links",
            };
          }

          const result = await createLinkAction({
            url,
            title: title || "",
            storyId,
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to add link",
            };
          }

          const newLink = result.data!;

          return {
            success: true,
            link: {
              id: newLink.id,
              url: newLink.url,
              title: newLink.title,
              createdAt: newLink.createdAt,
              updatedAt: newLink.updatedAt,
            },
            message: `Link "${newLink.title || newLink.url}" added successfully`,
          };
        }

        case "update-link": {
          if (!linkId || !url) {
            return {
              success: false,
              error: "Link ID and URL are required for updating links",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot update links",
            };
          }

          const result = await updateLinkAction(linkId, {
            url,
            title: title || "",
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to update link",
            };
          }

          return {
            success: true,
            message: "Link updated successfully",
          };
        }

        case "delete-link": {
          if (!linkId) {
            return {
              success: false,
              error: "Link ID is required for deleting links",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot delete links",
            };
          }

          const result = await deleteLinkAction(linkId);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to delete link",
            };
          }

          return {
            success: true,
            message: "Link deleted successfully",
          };
        }

        default:
          return {
            success: false,
            error: "Invalid link action",
          };
      }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "An error occurred",
      };
    }
  },
});
