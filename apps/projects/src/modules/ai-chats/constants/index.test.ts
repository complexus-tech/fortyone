/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { aiChatKeys } from "./index";

describe("aiChatKeys", () => {
  it("scopes every workspace-owned query beneath the workspace slug", () => {
    expect(aiChatKeys.lists("first")).toEqual([
      "ai-chats",
      "workspace",
      "first",
      "list",
    ]);
    expect(aiChatKeys.detail("first", "chat-1")).toEqual([
      "ai-chats",
      "workspace",
      "first",
      "detail",
      "chat-1",
    ]);
    expect(aiChatKeys.messages("first", "chat-1")).toEqual([
      "ai-chats",
      "workspace",
      "first",
      "detail",
      "chat-1",
      "messages",
    ]);
    expect(aiChatKeys.totalMessages("first")).toEqual([
      "ai-chats",
      "workspace",
      "first",
      "total-messages",
    ]);
    expect(aiChatKeys.memories("first")).toEqual([
      "ai-chats",
      "workspace",
      "first",
      "memories",
    ]);
    expect(aiChatKeys.memory("first", "memory-1")).toEqual([
      "ai-chats",
      "workspace",
      "first",
      "memories",
      "memory-1",
    ]);
  });

  it("does not share cache keys between workspaces", () => {
    expect(aiChatKeys.lists("first")).not.toEqual(aiChatKeys.lists("second"));
    expect(aiChatKeys.messages("first", "chat-1")).not.toEqual(
      aiChatKeys.messages("second", "chat-1"),
    );
    expect(aiChatKeys.memories("first")).not.toEqual(
      aiChatKeys.memories("second"),
    );
  });
});
