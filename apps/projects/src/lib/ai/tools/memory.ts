import { z } from "zod";
import { tool } from "ai";
import { createMemoryAction } from "@/modules/ai-chats/actions/create-memory";
import { updateMemoryAction } from "@/modules/ai-chats/actions/update-memory";
import { deleteMemoryAction } from "@/modules/ai-chats/actions/delete-memory";
import { getMemories } from "@/modules/ai-chats/queries/get-memory";
import { auth } from "@/auth";

export const listMemories = tool({
  description: "List all memories about the user.",
  inputSchema: z.object({}),
  execute: async () => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required",
        };
      }
      const memories = await getMemories(session);
      return {
        success: true,
        memories,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Failed to list memories",
      };
    }
  },
});

export const createMemory = tool({
  description:
    "Create a new memory about the user's preferences, role, or project context to improve future interactions.",
  inputSchema: z.object({
    content: z
      .string()
      .describe(
        "The content of the memory to save (e.g., 'The user is a senior frontend engineer') max length 200 words",
      ),
  }),
  execute: async ({ content }) => {
    try {
      const result = await createMemoryAction({ content });
      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message,
        };
      }
      return {
        success: true,
        memory: result.data,
        message: "Memory saved successfully",
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Failed to save memory",
      };
    }
  },
});

export const updateMemory = tool({
  description: "Update an existing memory about the user.",
  inputSchema: z.object({
    id: z.string().describe("The ID of the memory to update"),
    content: z.string().describe("The new content of the memory. max length 200 words"),
  }),
  execute: async ({ id, content }) => {
    try {
      const result = await updateMemoryAction(id, { content });
      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message,
        };
      }
      return {
        success: true,
        message: "Memory updated successfully",
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Failed to update memory",
      };
    }
  },
});

export const deleteMemory = tool({
  description: "Delete a memory about the user.",
  inputSchema: z.object({
    id: z.string().describe("The ID of the memory to delete"),
  }),
  execute: async ({ id }) => {
    try {
      const result = await deleteMemoryAction(id);
      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message,
        };
      }
      return {
        success: true,
        message: "Memory deleted successfully",
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Failed to delete memory",
      };
    }
  },
});
