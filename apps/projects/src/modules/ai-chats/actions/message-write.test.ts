/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { post } from "@/lib/http";
import {
  beginAiChatMessageWrite,
  finalizeAiChatMessageWrite,
  recoverMutationApprovalOutput,
} from "./message-write";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/http", () => ({ post: jest.fn() }));

const authMock = jest.mocked(auth);
const postMock = jest.mocked(post);
const reservation = {
  generation: 3,
  token: "c6cb8067-447f-41b0-bb66-2a09a0377993",
};

describe("message write actions", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue({ user: { id: "user-1" } } as never);
  });

  it("does not retry the non-idempotent begin reservation", async () => {
    postMock.mockResolvedValue({ data: reservation });

    await beginAiChatMessageWrite({
      chatId: "1234567890123456",
      messageId: "assistant-1",
      messages: [],
      operation: "regenerate",
      title: "Chat",
      workspaceSlug: "acme",
    });

    expect(postMock).toHaveBeenCalledWith(
      "chat-sessions/1234567890123456/message-writes/begin",
      {
        messageId: "assistant-1",
        messages: [],
        operation: "regenerate",
        title: "Chat",
      },
      expect.objectContaining({ workspaceSlug: "acme" }),
      { retry: 0, timeout: 10_000 },
    );
  });

  it("returns server-repaired request-safe messages with the reservation", async () => {
    const canonicalMessages = [
      {
        id: "assistant-1",
        parts: [
          {
            input: { title: "Launch" },
            output: { storyId: "story-1", success: true },
            state: "output-available",
            toolCallId: "call-1",
            type: "tool-createStory",
          },
        ],
        role: "assistant",
      },
    ];
    postMock.mockResolvedValue({
      data: { ...reservation, messages: canonicalMessages },
    });

    await expect(
      beginAiChatMessageWrite({
        chatId: "1234567890123456",
        messages: [],
        operation: "approval",
        title: "Chat",
        workspaceSlug: "acme",
      }),
    ).resolves.toEqual({ ...reservation, messages: canonicalMessages });
  });

  it("retries idempotent finalization once on timeout and retriable server failures", async () => {
    postMock.mockResolvedValue({ data: { applied: true } });

    await finalizeAiChatMessageWrite({
      chatId: "1234567890123456",
      messages: [],
      reservation,
      workspaceSlug: "acme",
    });

    expect(postMock).toHaveBeenCalledWith(
      "chat-sessions/1234567890123456/message-writes/finalize",
      { generation: 3, messages: [], token: reservation.token },
      expect.objectContaining({ workspaceSlug: "acme" }),
      {
        retry: {
          limit: 1,
          methods: ["post"],
          retryOnTimeout: true,
          statusCodes: [408, 500, 502, 503, 504],
        },
        timeout: 10_000,
      },
    );
  });

  it("uses the same bounded idempotent retry for targeted recovery", async () => {
    postMock.mockResolvedValue({ data: { applied: false } });

    await recoverMutationApprovalOutput({
      chatId: "1234567890123456",
      fingerprint: "a".repeat(64),
      toolCallId: "call/1",
      workspaceSlug: "acme",
    });

    expect(postMock).toHaveBeenCalledWith(
      "chat-sessions/1234567890123456/mutation-approvals/call%2F1/recover-output",
      { fingerprint: "a".repeat(64) },
      expect.objectContaining({ workspaceSlug: "acme" }),
      expect.objectContaining({
        retry: expect.objectContaining({
          limit: 1,
          retryOnTimeout: true,
        }),
      }),
    );
  });
});
