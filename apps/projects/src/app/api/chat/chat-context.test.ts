/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this docblock.
/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { ModelMessage, UIMessage } from "ai";
import {
  assertLatestUserTextWithinContextBudget,
  compactChatToolOutputs,
  compactUnknownChatToolOutputs,
  MAX_CHAT_CONTEXT_BYTES,
  omitHistoricalChatAttachments,
  pruneChatModelMessages,
  selectChatContextMessages,
} from "./chat-context";

const textMessage = (role: "assistant" | "user", index: number): UIMessage => ({
  id: `message-${index}`,
  parts: [{ text: `Message ${index}`, type: "text" }],
  role,
});

describe("chat context", () => {
  it("rejects one oversized latest text turn instead of bypassing the context budget", () => {
    const messages = [
      {
        id: "latest-user",
        parts: [
          { text: "x".repeat(MAX_CHAT_CONTEXT_BYTES + 1), type: "text" },
        ],
        role: "user",
      },
    ] satisfies UIMessage[];

    expect(() => {
      assertLatestUserTextWithinContextBudget(messages);
    }).toThrow("latest message text is too large");
  });

  it("does not count current attachments against the text-only context guard", () => {
    const messages = [
      {
        id: "latest-user",
        parts: [
          { text: "Analyze this image", type: "text" },
          {
            filename: "large.png",
            mediaType: "image/png",
            type: "file",
            url: `data:image/png;base64,${"x".repeat(MAX_CHAT_CONTEXT_BYTES)}`,
          },
        ],
        role: "user",
      },
    ] satisfies UIMessage[];

    expect(() => {
      assertLatestUserTextWithinContextBudget(messages);
    }).not.toThrow();
  });

  it("keeps the full semantic history and starts at a user boundary", () => {
    const messages = [
      textMessage("assistant", -1),
      ...Array.from({ length: 30 }, (_, index) =>
        textMessage(index % 2 === 0 ? "user" : "assistant", index),
      ),
    ];

    const contextMessages = selectChatContextMessages(messages);

    expect(contextMessages).toHaveLength(30);
    expect(contextMessages[0]?.id).toBe("message-0");
    expect(contextMessages.at(-1)?.id).toBe("message-29");
  });

  it("retains a large recent context without exceeding the provider safety budget", () => {
    const largeText = "x".repeat(MAX_CHAT_CONTEXT_BYTES / 2);
    const messages = [
      {
        id: "old-user",
        parts: [{ text: largeText, type: "text" }],
        role: "user",
      },
      {
        id: "old-assistant",
        parts: [{ text: largeText, type: "text" }],
        role: "assistant",
      },
      {
        id: "recent-user",
        parts: [{ text: largeText, type: "text" }],
        role: "user",
      },
    ] satisfies UIMessage[];

    const contextMessages = selectChatContextMessages(messages);

    expect(contextMessages[0]?.id).toBe("recent-user");
    expect(
      new TextEncoder().encode(JSON.stringify(contextMessages)).byteLength,
    ).toBeLessThanOrEqual(MAX_CHAT_CONTEXT_BYTES + 200);
  });

  it("budgets multi-byte Unicode by serialized bytes", () => {
    const largeText = "🗓️".repeat(MAX_CHAT_CONTEXT_BYTES / 12);
    const messages = [
      {
        id: "old-user",
        parts: [{ text: largeText, type: "text" }],
        role: "user",
      },
      {
        id: "old-assistant",
        parts: [{ text: largeText, type: "text" }],
        role: "assistant",
      },
      {
        id: "recent-user",
        parts: [{ text: largeText, type: "text" }],
        role: "user",
      },
    ] satisfies UIMessage[];

    const contextMessages = selectChatContextMessages(messages);
    const serializedBytes = new TextEncoder().encode(
      JSON.stringify(contextMessages),
    ).byteLength;

    expect(contextMessages[0]?.id).toBe("recent-user");
    expect(serializedBytes).toBeLessThanOrEqual(MAX_CHAT_CONTEXT_BYTES + 200);
  });

  it("omits historical attachments without dropping conversation text", () => {
    const messages = [
      {
        id: "historical-user-message",
        parts: [
          { text: "Please analyze this file", type: "text" },
          {
            filename: "historical.png",
            mediaType: "image/png",
            type: "file",
            url: "data:image/png;base64,historical-payload",
          },
        ],
        role: "user",
      },
      textMessage("assistant", 1),
      {
        id: "current-user-message",
        parts: [
          { text: "Compare it with this one", type: "text" },
          {
            filename: "current.png",
            mediaType: "image/png",
            type: "file",
            url: "data:image/png;base64,current-payload",
          },
        ],
        role: "user",
      },
    ] satisfies UIMessage[];

    const preparedMessages = omitHistoricalChatAttachments(messages);
    const serialized = JSON.stringify(preparedMessages);

    expect(serialized).toContain("Please analyze this file");
    expect(serialized).toContain(
      "[Historical attachment omitted from model context]",
    );
    expect(serialized).not.toContain("historical-payload");
    expect(serialized).toContain("current-payload");
    expect(messages[0]?.parts[1]).toHaveProperty(
      "url",
      "data:image/png;base64,historical-payload",
    );
  });

  it("retains exact tool receipts and provider-linked reasoning", () => {
    const messages = [
      {
        role: "assistant",
        content: [
          {
            type: "text",
            text: "Earlier semantic decision that must remain available",
          },
          {
            type: "reasoning",
            text: "Historical private reasoning",
          },
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

    expect(JSON.stringify(prunedMessages)).toContain("large");
    expect(JSON.stringify(prunedMessages)).toContain("old-call");
    expect(JSON.stringify(prunedMessages)).toContain(
      "Historical private reasoning",
    );
    expect(JSON.stringify(prunedMessages)).toContain(
      "Earlier semantic decision that must remain available",
    );
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

  it("compacts only unregistered historical tool outputs before conversion", () => {
    const messages = [
      {
        id: "assistant-message",
        parts: [
          {
            input: {},
            output: {
              contentHTML: `<p>${"legacy".repeat(500)}</p>`,
              success: true,
            },
            state: "output-available",
            toolCallId: "legacy-call",
            type: "tool-removedLegacyTool",
          },
          {
            input: {},
            output: {
              installUrl:
                "https://github.com/apps/fortyone/installations/new?state=signed",
              success: true,
            },
            state: "output-available",
            toolCallId: "registered-call",
            type: "tool-createGitHubInstallSessionTool",
          },
        ],
        role: "assistant",
      },
    ] as unknown as UIMessage[];

    const prepared = compactUnknownChatToolOutputs(
      messages,
      new Set(["createGitHubInstallSessionTool"]),
    );

    expect(prepared[0]?.parts[0]).not.toHaveProperty("output.contentHTML");
    expect(prepared[0]?.parts[1]).toHaveProperty(
      "output.installUrl",
      "https://github.com/apps/fortyone/installations/new?state=signed",
    );
  });
});
