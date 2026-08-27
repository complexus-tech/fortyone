/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { post } from "@/lib/http";
import { getLatestAiChatAssistantMessage } from "../queries/get-latest-ai-chat-assistant-message";
import {
  claimMutationApprovalExecution,
  completeMutationApprovalExecution,
  failMutationApprovalExecution,
  getPersistedMutationApprovalMessage,
  startMutationApprovalExecution,
} from "./mutation-approval-execution";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/http", () => ({ post: jest.fn() }));
jest.mock("../queries/get-latest-ai-chat-assistant-message", () => ({
  getLatestAiChatAssistantMessage: jest.fn(),
}));

const mockAuth = jest.mocked(auth);
const mockPost = jest.mocked(post);
const mockGetLatestAssistantMessage = jest.mocked(
  getLatestAiChatAssistantMessage,
);
const LEASE_TOKEN = "55a2c9f1-51d9-4c7f-b77b-e4c2f92439ba";

describe("mutation approval execution actions", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockAuth.mockResolvedValue({ user: { id: "user-1" } } as never);
  });

  it("claims an owner-scoped approval using an encoded tool call id", async () => {
    mockPost.mockResolvedValue({
      data: { leaseToken: LEASE_TOKEN, state: "claimed" },
    });

    await expect(
      claimMutationApprovalExecution({
        chatId: "1234567890123456",
        fingerprint: "a".repeat(64),
        toolCallId: "call/with spaces",
        workspaceSlug: "acme",
      }),
    ).resolves.toEqual({ leaseToken: LEASE_TOKEN, state: "claimed" });

    expect(mockPost).toHaveBeenCalledWith(
      "chat-sessions/1234567890123456/mutation-approvals/call%2Fwith%20spaces/claim",
      { fingerprint: "a".repeat(64) },
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
  });

  it("completes with JSON null when a tool returns undefined", async () => {
    mockPost.mockResolvedValue({
      data: { output: null, state: "completed" },
    });

    await expect(
      completeMutationApprovalExecution({
        chatId: "1234567890123456",
        fingerprint: "b".repeat(64),
        leaseToken: LEASE_TOKEN,
        output: undefined,
        toolCallId: "call-1",
        workspaceSlug: "acme",
      }),
    ).resolves.toEqual({ output: null, state: "completed" });

    expect(mockPost).toHaveBeenCalledWith(
      expect.stringContaining("/complete"),
      {
        fingerprint: "b".repeat(64),
        leaseToken: LEASE_TOKEN,
        output: null,
      },
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
  });

  it("starts only the exact leased execution", async () => {
    mockPost.mockResolvedValue({ data: { state: "started" } });

    await expect(
      startMutationApprovalExecution({
        chatId: "1234567890123456",
        fingerprint: "d".repeat(64),
        leaseToken: LEASE_TOKEN,
        toolCallId: "call-1",
        workspaceSlug: "acme",
      }),
    ).resolves.toEqual({ state: "started" });

    expect(mockPost).toHaveBeenCalledWith(
      expect.stringContaining("/start"),
      { fingerprint: "d".repeat(64), leaseToken: LEASE_TOKEN },
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
  });

  it("terminally records an ambiguous completion", async () => {
    mockPost.mockResolvedValue({
      data: {
        failureCode: "completion_persistence_uncertain",
        state: "failed_uncertain",
      },
    });

    await expect(
      failMutationApprovalExecution({
        chatId: "1234567890123456",
        failureCode: "completion_persistence_uncertain",
        fingerprint: "e".repeat(64),
        leaseToken: LEASE_TOKEN,
        toolCallId: "call-1",
        workspaceSlug: "acme",
      }),
    ).resolves.toEqual({
      failureCode: "completion_persistence_uncertain",
      state: "failed_uncertain",
    });

    expect(mockPost).toHaveBeenCalledWith(
      expect.stringContaining("/fail"),
      {
        failureCode: "completion_persistence_uncertain",
        fingerprint: "e".repeat(64),
        leaseToken: LEASE_TOKEN,
      },
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
  });

  it("rejects a claimed response without a lease", async () => {
    mockPost.mockResolvedValue({ data: { state: "claimed" } });

    await expect(
      claimMutationApprovalExecution({
        chatId: "1234567890123456",
        fingerprint: "f".repeat(64),
        toolCallId: "call-1",
        workspaceSlug: "acme",
      }),
    ).rejects.toThrow("no execution lease");
  });

  it("fetches persisted messages through the owner-scoped chat endpoint", async () => {
    const messages = [{ id: "assistant-1", parts: [], role: "assistant" }];
    mockGetLatestAssistantMessage.mockResolvedValue(messages[0] as never);

    await expect(
      getPersistedMutationApprovalMessage("1234567890123456", "acme"),
    ).resolves.toBe(messages[0]);
    expect(mockGetLatestAssistantMessage).toHaveBeenCalledWith(
      expect.objectContaining({ workspaceSlug: "acme" }),
      "1234567890123456",
    );
  });

  it("fails before an API call without an authenticated user", async () => {
    mockAuth.mockResolvedValue(null);

    await expect(
      claimMutationApprovalExecution({
        chatId: "1234567890123456",
        fingerprint: "c".repeat(64),
        toolCallId: "call-1",
        workspaceSlug: "acme",
      }),
    ).rejects.toThrow("authentication is required");
    expect(mockPost).not.toHaveBeenCalled();
  });
});
