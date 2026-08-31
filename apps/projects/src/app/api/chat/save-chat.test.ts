/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { UIMessage } from "ai";
import {
  beginAiChatMessageWrite,
  finalizeAiChatMessageWrite,
} from "@/modules/ai-chats/actions/message-write";
import { beginChatWrite, saveChat } from "./save-chat";

jest.mock("@/modules/ai-chats/actions/message-write", () => ({
  beginAiChatMessageWrite: jest.fn(),
  finalizeAiChatMessageWrite: jest.fn(),
}));

const beginWriteMock = jest.mocked(beginAiChatMessageWrite);
const finalizeWriteMock = jest.mocked(finalizeAiChatMessageWrite);
const reservation = {
  generation: 7,
  token: "55a2c9f1-51d9-4c7f-b77b-e4c2f92439ba",
};
const messages = [
  {
    id: "user-1",
    parts: [
      {
        text: "Create the launch checklist stories for Product.",
        type: "text",
      },
    ],
    role: "user",
  },
  {
    id: "assistant-1",
    parts: [{ text: "I prepared the stories for approval.", type: "text" }],
    role: "assistant",
  },
] satisfies UIMessage[];

describe("chat write persistence", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    beginWriteMock.mockResolvedValue(reservation);
    finalizeWriteMock.mockResolvedValue({ applied: true });
  });

  it("captures a server reservation with deterministic title and regeneration target", async () => {
    await expect(
      beginChatWrite({
        id: "chat-12345678901",
        messageId: "assistant-1",
        messages,
        operation: "regenerate",
        workspaceSlug: "acme",
      }),
    ).resolves.toEqual(reservation);

    expect(beginWriteMock).toHaveBeenCalledWith({
      chatId: "chat-12345678901",
      messageId: "assistant-1",
      messages,
      operation: "regenerate",
      title: "Create the launch checklist stories for Product.",
      workspaceSlug: "acme",
    });
  });

  it("finalizes only through the captured token and generation", async () => {
    await expect(
      saveChat({
        id: "chat-12345678901",
        messages,
        reservation,
        workspaceSlug: "acme",
      }),
    ).resolves.toEqual({ applied: true });

    expect(finalizeWriteMock).toHaveBeenCalledWith({
      chatId: "chat-12345678901",
      messages,
      reservation,
      workspaceSlug: "acme",
    });
  });

  it("surfaces rejected finalization without logging transcript content", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    finalizeWriteMock.mockRejectedValue(
      Object.assign(new Error("Database unavailable"), { code: "timeout" }),
    );

    await expect(
      saveChat({
        id: "chat-12345678901",
        messages,
        reservation,
        workspaceSlug: "acme",
      }),
    ).rejects.toThrow("Database unavailable");

    expect(consoleError).toHaveBeenCalledWith(
      "[chat/save] Failed to finalize Maya conversation write",
      JSON.stringify({
        chatId: "chat-12345678901",
        generation: 7,
        workspaceSlug: "acme",
        error: {
          codes: ["timeout"],
          errorType: "Error",
          statuses: [],
        },
      }),
    );
    consoleError.mockRestore();
  });
});
