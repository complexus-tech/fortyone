/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this docblock.
/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { ModelMessage, UIMessage } from "ai";
import {
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
});
