/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this docblock.
/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { ModelMessage, UIMessage } from "ai";
import {
  compactChatToolOutputs,
  MAX_CHAT_CONTEXT_MESSAGES,
  pruneChatModelMessages,
  selectRecentChatMessages,
} from "./chat-context";

const textMessage = (role: "assistant" | "user", index: number): UIMessage => ({
  id: `message-${index}`,
  parts: [{ text: `Message ${index}`, type: "text" }],
  role,
});

describe("chat context", () => {
  it("keeps only recent UI messages and starts at a user boundary", () => {
    const messages = Array.from({ length: 30 }, (_, index) =>
      textMessage(index % 2 === 0 ? "user" : "assistant", index),
    );

    const recentMessages = selectRecentChatMessages(messages);

    expect(recentMessages.length).toBeLessThanOrEqual(
      MAX_CHAT_CONTEXT_MESSAGES,
    );
    expect(recentMessages[0]?.role).toBe("user");
    expect(recentMessages.at(-1)?.id).toBe("message-29");
  });

  it("removes old tool payloads while retaining recent conversation text", () => {
    const messages = [
      {
        role: "assistant",
        content: [
          {
            type: "tool-call",
            toolCallId: "old-call",
            toolName: "search",
            input: { query: "old data" },
          },
        ],
      },
      {
        role: "tool",
        content: [
          {
            type: "tool-result",
            toolCallId: "old-call",
            toolName: "search",
            output: { type: "json", value: { large: "payload" } },
          },
        ],
      },
      ...Array.from({ length: 9 }, (_, index) => ({
        role: "user" as const,
        content: `Recent message ${index}`,
      })),
    ] as ModelMessage[];

    const prunedMessages = pruneChatModelMessages(messages);

    expect(JSON.stringify(prunedMessages)).not.toContain("large");
    expect(JSON.stringify(prunedMessages)).toContain("Recent message 8");
  });

  it("compacts recent tool outputs before converting chat history", () => {
    const messages = [
      {
        id: "assistant-message",
        parts: [
          {
            input: {},
            output: {
              content: "x".repeat(2000),
              contentHTML: `<p>${"x".repeat(2000)}</p>`,
              success: true,
            },
            state: "output-available",
            type: "tool-getDocumentDetailsTool",
          },
        ],
        role: "assistant",
      },
    ] as unknown as UIMessage[];

    const compacted = compactChatToolOutputs(messages);
    const serialized = JSON.stringify(compacted);

    expect(serialized).not.toContain("contentHTML");
    expect(serialized.length).toBeLessThan(1500);
    expect(messages[0]?.parts[0]).toHaveProperty("output.contentHTML");
  });
});
