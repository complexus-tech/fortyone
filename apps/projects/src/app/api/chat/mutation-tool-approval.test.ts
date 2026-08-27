/* global beforeAll, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import {
  ReadableStream as NodeReadableStream,
  TextDecoderStream as NodeTextDecoderStream,
  TextEncoderStream as NodeTextEncoderStream,
  TransformStream as NodeTransformStream,
  WritableStream as NodeWritableStream,
} from "node:stream/web";
import {
  TextDecoder as NodeTextDecoder,
  TextEncoder as NodeTextEncoder,
} from "node:util";
import type * as ZodModule from "zod";
import { ApiError } from "api-client";
import { tools } from "@/lib/ai/tools";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import {
  claimMutationApprovalExecution,
  completeMutationApprovalExecution,
  failMutationApprovalExecution,
  startMutationApprovalExecution,
} from "@/modules/ai-chats/actions/mutation-approval-execution";
import { recoverMutationApprovalOutput } from "@/modules/ai-chats/actions/message-write";
import { getApiError } from "@/utils";
import { beginChatWrite, saveChat } from "./save-chat";
import type * as MutationApprovalModule from "./mutation-tool-approval";

globalThis.ReadableStream =
  NodeReadableStream as typeof globalThis.ReadableStream;
globalThis.TransformStream =
  NodeTransformStream as typeof globalThis.TransformStream;
globalThis.WritableStream =
  NodeWritableStream as typeof globalThis.WritableStream;
globalThis.TextDecoder = NodeTextDecoder as typeof globalThis.TextDecoder;
globalThis.TextDecoderStream =
  NodeTextDecoderStream as typeof globalThis.TextDecoderStream;
globalThis.TextEncoder = NodeTextEncoder as typeof globalThis.TextEncoder;
globalThis.TextEncoderStream =
  NodeTextEncoderStream as typeof globalThis.TextEncoderStream;
globalThis.structuredClone = <Value>(value: Value) =>
  JSON.parse(JSON.stringify(value)) as Value;

jest.mock("@/lib/ai/tools", () => {
  const { z } = jest.requireActual<typeof ZodModule>("zod");

  return {
    tools: {
      comments: {
        execute: jest.fn(async () => ({ success: true })),
        inputSchema: z.object({
          action: z.enum(["add-comment", "list-comments"]),
          content: z.string().optional(),
          storyId: z.string(),
        }),
      },
      createGitHubInstallSessionTool: {
        execute: jest.fn(async () => ({ success: true })),
        inputSchema: z.object({}),
      },
      createStory: {
        execute: jest.fn(async (input: { title: string }) => ({
          message: `Story "${input.title}" created successfully.`,
          success: true,
        })),
        inputSchema: z.object({
          teamId: z.string(),
          title: z.string(),
        }),
      },
      updateStory: {
        execute: jest.fn(async () => ({ success: true })),
        inputSchema: z.object({
          confirmed: z.boolean().optional(),
          description: z.string().nullable().optional(),
          dueDate: z.string().optional(),
          storyId: z.string(),
          title: z.string(),
        }),
      },
    },
  };
});

jest.mock("api-client", () => ({
  ApiError: class ApiError extends Error {
    data: unknown;
    status: number;

    constructor(message: string, status: number, data: unknown) {
      super(message);
      this.data = data;
      this.status = status;
    }
  },
}));

jest.mock("./save-chat", () => ({
  beginChatWrite: jest.fn(),
  saveChat: jest.fn(),
}));
jest.mock("@/modules/ai-chats/actions/message-write", () => ({
  recoverMutationApprovalOutput: jest.fn(),
}));
jest.mock("@/modules/ai-chats/actions/mutation-approval-execution", () => ({
  claimMutationApprovalExecution: jest.fn(),
  completeMutationApprovalExecution: jest.fn(),
  failMutationApprovalExecution: jest.fn(),
  startMutationApprovalExecution: jest.fn(),
}));

type ExecuteMock = ReturnType<typeof jest.fn>;

const mockCommentsExecute = tools.comments.execute as unknown as ExecuteMock;
const mockCreateStoryExecute = tools.createStory
  .execute as unknown as ExecuteMock;
const mockGitHubSetupExecute = tools.createGitHubInstallSessionTool
  .execute as unknown as ExecuteMock;
const mockUpdateStoryExecute = tools.updateStory
  .execute as unknown as ExecuteMock;
const mockBeginChatWrite = jest.mocked(beginChatWrite);
const mockSaveChat = jest.mocked(saveChat);
const mockRecoverMutationApprovalOutput = jest.mocked(
  recoverMutationApprovalOutput,
);
const mockClaimMutationApproval = jest.mocked(claimMutationApprovalExecution);
const mockCompleteMutationApproval = jest.mocked(
  completeMutationApprovalExecution,
);
const mockFailMutationApproval = jest.mocked(failMutationApprovalExecution);
const mockStartMutationApproval = jest.mocked(startMutationApprovalExecution);
const LEASE_TOKEN = "55a2c9f1-51d9-4c7f-b77b-e4c2f92439ba";
const WRITE_RESERVATION = {
  generation: 1,
  token: "92158abf-fcb4-4fb8-84fb-9d553a50ed5e",
};

let approvalModule: typeof MutationApprovalModule;

const approvalPart = ({
  approved = true,
  input,
  toolCallId = "call-1",
  toolName = "createStory",
}: {
  approved?: boolean;
  input: unknown;
  toolCallId?: string;
  toolName?: string;
}) => ({
  approval: { approved, id: `approval-${toolCallId}` },
  input,
  state: "approval-responded",
  toolCallId,
  type: `tool-${toolName}`,
});

const approvalMessage = (...parts: ReturnType<typeof approvalPart>[]) =>
  ({
    id: "assistant-1",
    parts,
    role: "assistant",
  }) as unknown as MayaUIMessage;

const createResponse = ({
  abortSignal = new AbortController().signal,
  chatId = "chat-1",
  messages,
  userId = "user-1",
}: {
  abortSignal?: AbortSignal;
  chatId?: string;
  messages: MayaUIMessage[];
  userId?: string;
}) => {
  const response = approvalModule.createMutationToolApprovalResponse({
    abortSignal,
    chatId,
    messages,
    userId,
    workspaceSlug: "acme",
  });
  if (!response || response instanceof Response) return response;

  // Production awaits the reservation-backed response before returning it.
  // Keep existing stream assertions concise while preserving the synchronous
  // undefined result for messages that do not contain mutation approvals.
  return {
    text: async () => (await response).text(),
  } as Response;
};

describe("mutation tool approval response", () => {
  beforeAll(async () => {
    const { Response: EdgeResponse } = await import(
      "next/dist/compiled/@edge-runtime/primitives/fetch"
    );
    globalThis.Response = EdgeResponse;
    approvalModule = await import("./mutation-tool-approval");
  });

  beforeEach(() => {
    jest.clearAllMocks();
    approvalModule.resetMutationApprovalCacheForTests();
    mockBeginChatWrite.mockResolvedValue(WRITE_RESERVATION);
    mockSaveChat.mockResolvedValue({ applied: true });
    mockRecoverMutationApprovalOutput.mockResolvedValue({ applied: true });
    const durableExecutions = new Map<
      string,
      {
        fingerprint: string;
        failureCode?: string;
        output?: unknown;
        state: "completed" | "failed_uncertain" | "in_progress";
      }
    >();
    mockClaimMutationApproval.mockImplementation(async (input) => {
      const key = `${input.chatId}:${input.toolCallId}`;
      const existing = durableExecutions.get(key);
      if (!existing) {
        durableExecutions.set(key, {
          fingerprint: input.fingerprint,
          state: "in_progress",
        });
        return { leaseToken: LEASE_TOKEN, state: "claimed" };
      }
      if (existing.fingerprint !== input.fingerprint) {
        throw new Error("fingerprint conflict");
      }
      return existing;
    });
    mockStartMutationApproval.mockResolvedValue({ state: "started" });
    mockCompleteMutationApproval.mockImplementation(async (input) => {
      const execution = {
        fingerprint: input.fingerprint,
        output: input.output,
        state: "completed" as const,
      };
      durableExecutions.set(`${input.chatId}:${input.toolCallId}`, execution);
      return execution;
    });
    mockFailMutationApproval.mockImplementation(async (input) => {
      const key = `${input.chatId}:${input.toolCallId}`;
      const existing = durableExecutions.get(key);
      if (existing?.state === "completed") return existing;
      const execution = {
        failureCode: input.failureCode,
        fingerprint: input.fingerprint,
        state: "failed_uncertain" as const,
      };
      durableExecutions.set(key, execution);
      return execution;
    });
    mockCreateStoryExecute.mockImplementation(async (input) => ({
      message: `Story "${input.title}" created successfully.`,
      success: true,
    }));
    mockUpdateStoryExecute.mockImplementation(async () => ({ success: true }));
    mockCommentsExecute.mockImplementation(async () => ({ success: true }));
  });

  it("executes the exact approved payload without another model call", async () => {
    const input = {
      teamId: "team-1",
      title: "Add onboarding checklist",
    };
    const response = createResponse({
      messages: [approvalMessage(approvalPart({ input }))],
    });

    expect(response).toBeDefined();
    const streamBody = await response!.text();

    expect(mockCreateStoryExecute).toHaveBeenCalledWith(
      input,
      expect.objectContaining({
        experimental_context: { chatId: "chat-1", workspaceSlug: "acme" },
        toolCallId: "call-1",
      }),
    );
    expect(mockClaimMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({
        chatId: "chat-1",
        fingerprint: expect.stringMatching(/^[0-9a-f]{64}$/),
        toolCallId: "call-1",
        workspaceSlug: "acme",
      }),
    );
    expect(mockStartMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({
        leaseToken: LEASE_TOKEN,
        toolCallId: "call-1",
      }),
    );
    expect(mockCompleteMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({
        leaseToken: LEASE_TOKEN,
        output: expect.objectContaining({ success: true }),
      }),
    );
    expect(streamBody).toContain("tool-output-available");
    expect(streamBody).toContain("created successfully");
    expect(mockSaveChat).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "chat-1",
        reservation: WRITE_RESERVATION,
        workspaceSlug: "acme",
      }),
    );
  });

  it("executes a server-prepared safe retry exactly once through the normal lease path", async () => {
    const toolCallId = "call-safe-retry";
    const input = {
      teamId: "team-1",
      title: "Recover the original story creation",
    };
    const submittedMessage = approvalMessage(
      approvalPart({ input, toolCallId }),
    );
    const canonicalMessage = approvalMessage(
      approvalPart({ input, toolCallId }),
    );
    const canonicalReservation = {
      ...WRITE_RESERVATION,
      messages: [canonicalMessage],
    };
    mockBeginChatWrite.mockResolvedValue(canonicalReservation);

    const response = createResponse({ messages: [submittedMessage] });
    const body = await response!.text();

    expect(mockClaimMutationApproval).toHaveBeenCalledTimes(1);
    const claimedIdentity = mockClaimMutationApproval.mock.calls[0][0];
    expect(claimedIdentity).toEqual(
      expect.objectContaining({
        chatId: "chat-1",
        fingerprint: expect.stringMatching(/^[0-9a-f]{64}$/),
        toolCallId,
        workspaceSlug: "acme",
      }),
    );
    expect(mockStartMutationApproval).toHaveBeenCalledTimes(1);
    expect(mockStartMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({
        fingerprint: claimedIdentity.fingerprint,
        toolCallId,
      }),
    );
    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(mockCompleteMutationApproval).toHaveBeenCalledTimes(1);
    expect(mockCompleteMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({
        fingerprint: claimedIdentity.fingerprint,
        toolCallId,
      }),
    );
    expect(body).toContain("tool-output-available");
    expect(body).toContain("created successfully");
  });

  it("never claims or starts a mutation when approval reservation fails", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    mockBeginChatWrite.mockRejectedValue(new Error("approval conflict"));
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Stale approval" },
          }),
        ),
      ],
    });

    await expect(response!.text()).rejects.toThrow("approval conflict");

    expect(mockClaimMutationApproval).not.toHaveBeenCalled();
    expect(mockStartMutationApproval).not.toHaveBeenCalled();
    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    consoleError.mockRestore();
  });

  it("recovers a durable output when reservation finalization is superseded", async () => {
    mockSaveChat.mockResolvedValue({ applied: false });
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Recover output" },
          }),
        ),
      ],
    });

    await response!.text();

    expect(mockRecoverMutationApprovalOutput).toHaveBeenCalledWith({
      chatId: "chat-1",
      fingerprint: expect.stringMatching(/^[0-9a-f]{64}$/),
      toolCallId: "call-1",
      workspaceSlug: "acme",
    });
    expect(mockFailMutationApproval).not.toHaveBeenCalled();
  });

  it("recovers completed output after a finalization transport failure without quarantining it", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    mockSaveChat.mockRejectedValue(new Error("finalize response lost"));
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Recover after failure" },
          }),
        ),
      ],
    });

    await response!.text();

    expect(mockRecoverMutationApprovalOutput).toHaveBeenCalledTimes(1);
    expect(mockFailMutationApproval).not.toHaveBeenCalled();
    consoleError.mockRestore();
  });

  it("executes only the remaining approval after a partial multi-approval output", async () => {
    const completedPart = {
      ...approvalPart({
        input: { teamId: "team-1", title: "Already completed" },
        toolCallId: "call-a",
      }),
      output: { storyId: "story-a", success: true },
      state: "output-available",
    };
    const remainingPart = approvalPart({
      input: { teamId: "team-1", title: "Still pending" },
      toolCallId: "call-b",
    });
    const response = createResponse({
      messages: [
        {
          id: "assistant-1",
          parts: [completedPart, remainingPart],
          role: "assistant",
        } as unknown as MayaUIMessage,
      ],
    });

    await response!.text();

    expect(mockClaimMutationApproval).toHaveBeenCalledTimes(1);
    expect(mockClaimMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({ toolCallId: "call-b" }),
    );
    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(mockCreateStoryExecute).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Still pending" }),
      expect.objectContaining({ toolCallId: "call-b" }),
    );
  });

  it("injects confirmed only for a legacy mutation after validation", async () => {
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: {
              confirmed: false,
              storyId: "story-1",
              title: "Updated title",
            },
            toolName: "updateStory",
          }),
        ),
      ],
    });

    await response!.text();

    expect(mockUpdateStoryExecute).toHaveBeenCalledWith(
      {
        confirmed: true,
        storyId: "story-1",
        title: "Updated title",
      },
      expect.objectContaining({ toolCallId: "call-1" }),
    );
  });

  it("preserves intentional null clears while removing strict-schema placeholder nulls", async () => {
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: {
              confirmed: false,
              description: null,
              dueDate: null,
              storyId: "story-1",
              title: "Updated title",
            },
            toolName: "updateStory",
          }),
        ),
      ],
    });

    await response!.text();

    expect(mockUpdateStoryExecute).toHaveBeenCalledWith(
      {
        confirmed: true,
        description: null,
        storyId: "story-1",
        title: "Updated title",
      },
      expect.objectContaining({ toolCallId: "call-1" }),
    );
  });

  it("executes an approved dynamic mutation action without altering its input", async () => {
    const input = {
      action: "add-comment" as const,
      content: "Looks good",
      storyId: "story-1",
    };
    const response = createResponse({
      messages: [
        approvalMessage(approvalPart({ input, toolName: "comments" })),
      ],
    });

    await response!.text();

    expect(mockCommentsExecute).toHaveBeenCalledWith(
      input,
      expect.objectContaining({ toolCallId: "call-1" }),
    );
  });

  it("serializes multiple approved mutations in their displayed order", async () => {
    const events: string[] = [];
    mockCreateStoryExecute.mockImplementation(async () => {
      events.push("create:start");
      await Promise.resolve();
      events.push("create:end");
      return { success: true };
    });
    mockUpdateStoryExecute.mockImplementation(async () => {
      events.push("update:start");
      await Promise.resolve();
      events.push("update:end");
      return { success: true };
    });

    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "New story" },
            toolCallId: "call-create",
          }),
          approvalPart({
            input: { storyId: "story-1", title: "Updated story" },
            toolCallId: "call-update",
            toolName: "updateStory",
          }),
        ),
      ],
    });

    await response!.text();

    expect(events).toEqual([
      "create:start",
      "create:end",
      "update:start",
      "update:end",
    ]);
  });

  it("halts later approved mutations after an earlier terminal failure", async () => {
    mockCreateStoryExecute.mockImplementation(async () => ({
      error: "The team no longer exists.",
      success: false,
    }));
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "First mutation" },
            toolCallId: "call-create",
          }),
          approvalPart({
            input: { storyId: "story-1", title: "Must not run" },
            toolCallId: "call-update",
            toolName: "updateStory",
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(mockUpdateStoryExecute).not.toHaveBeenCalled();
    expect(mockClaimMutationApproval).toHaveBeenCalledTimes(1);
    expect(body).toContain("earlier approved change was unresolved");
  });

  it("uses the server-repaired transcript as the approval stream base", async () => {
    const submitted = approvalMessage(
      approvalPart({
        input: { teamId: "team-1", title: "Already durable" },
      }),
    );
    const canonicalPart = approvalPart({
      input: { teamId: "team-1", title: "Already durable" },
    });
    Object.assign(canonicalPart, {
      output: { storyId: "story-1", success: true },
      state: "output-available",
    });
    const canonical = approvalMessage(canonicalPart);
    const canonicalReservation = {
      ...WRITE_RESERVATION,
      messages: [canonical],
    };
    mockBeginChatWrite.mockResolvedValue(canonicalReservation);

    await createResponse({ messages: [submitted] })!.text();

    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    expect(mockClaimMutationApproval).not.toHaveBeenCalled();
    expect(mockSaveChat).toHaveBeenCalledWith({
      id: "chat-1",
      messages: [canonical],
      reservation: canonicalReservation,
      workspaceSlug: "acme",
    });
  });

  it("never reopens a later approval that the server skipped after a durable failure", async () => {
    const submitted = approvalMessage(
      approvalPart({
        input: { teamId: "team-1", title: "Failed first" },
        toolCallId: "call-a",
      }),
      approvalPart({
        input: { storyId: "story-1", title: "Must remain skipped" },
        toolCallId: "call-b",
        toolName: "updateStory",
      }),
    );
    const failedPart = approvalPart({
      input: { teamId: "team-1", title: "Failed first" },
      toolCallId: "call-a",
    });
    Object.assign(failedPart, {
      output: { error: "Team not found", success: false },
      state: "output-available",
    });
    const skippedPart = approvalPart({
      input: { storyId: "story-1", title: "Must remain skipped" },
      toolCallId: "call-b",
      toolName: "updateStory",
    });
    Object.assign(skippedPart, {
      output: {
        error:
          "Maya did not run this approved change because an earlier approved change was unresolved. Review the earlier result, then ask Maya to prepare this change again.",
        success: false,
      },
      state: "output-available",
    });
    const canonical = approvalMessage(failedPart, skippedPart);
    mockBeginChatWrite.mockResolvedValue({
      ...WRITE_RESERVATION,
      messages: [canonical],
    });

    const body = await createResponse({ messages: [submitted] })!.text();

    expect(mockClaimMutationApproval).not.toHaveBeenCalled();
    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    expect(mockUpdateStoryExecute).not.toHaveBeenCalled();
    expect(body).toContain("earlier approved change was unresolved");
  });

  it("records a denial without executing the tool", async () => {
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            approved: false,
            input: { teamId: "team-1", title: "Do not create this" },
          }),
        ),
      ],
    });

    const streamBody = await response!.text();

    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    expect(mockClaimMutationApproval).not.toHaveBeenCalled();
    expect(streamBody).toContain("tool-output-denied");
    expect(mockSaveChat).toHaveBeenCalled();
  });

  it("does not require a ledger recovery when a denial finalization is superseded", async () => {
    mockSaveChat.mockResolvedValue({ applied: false });
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            approved: false,
            input: { teamId: "team-1", title: "Denied durably at begin" },
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(body).toContain("tool-output-denied");
    expect(mockBeginChatWrite).toHaveBeenCalledWith(
      expect.objectContaining({ operation: "approval" }),
    );
    expect(mockRecoverMutationApprovalOutput).not.toHaveBeenCalled();
    expect(mockClaimMutationApproval).not.toHaveBeenCalled();
  });

  it("reuses an identical approval result without replaying the mutation", async () => {
    const messages = [
      approvalMessage(
        approvalPart({
          input: { teamId: "team-1", title: "Only once" },
        }),
      ),
    ];

    const firstBody = await createResponse({ messages })!.text();
    const replayBody = await createResponse({ messages })!.text();

    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(mockClaimMutationApproval).toHaveBeenCalledTimes(2);
    expect(mockBeginChatWrite).toHaveBeenCalledTimes(2);
    expect(firstBody).toContain("tool-output-available");
    expect(replayBody).toContain("tool-output-available");
  });

  it("coalesces concurrent approval requests in the same process", async () => {
    let releaseExecution: (() => void) | undefined;
    let markExecutionStarted: (() => void) | undefined;
    const executionStarted = new Promise<void>((resolve) => {
      markExecutionStarted = resolve;
    });
    mockCreateStoryExecute.mockImplementation(
      async () =>
        new Promise((resolve) => {
          markExecutionStarted?.();
          releaseExecution = () => {
            resolve({ success: true });
          };
        }),
    );
    const messages = [
      approvalMessage(
        approvalPart({
          input: { teamId: "team-1", title: "Coalesced result" },
        }),
      ),
    ];

    const firstBody = createResponse({ messages, userId: "user-1" })!.text();
    const secondBody = createResponse({ messages, userId: "user-1" })!.text();
    await executionStarted;
    releaseExecution!();
    await Promise.all([firstBody, secondBody]);

    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(mockClaimMutationApproval).toHaveBeenCalledTimes(1);
  });

  it("rejects a changed payload for an existing approval identity", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    await createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Original" },
          }),
        ),
      ],
    })!.text();

    const staleBody = await createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Changed" },
          }),
        ),
      ],
    })!.text();

    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(staleBody).toContain("could not safely confirm");
    consoleError.mockRestore();
  });

  it("rejects invalid input without executing the tool", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1" },
          }),
        ),
      ],
    });

    const streamBody = await response!.text();

    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    expect(mockClaimMutationApproval).not.toHaveBeenCalled();
    expect(streamBody).toContain("was not executed");
    consoleError.mockRestore();
  });

  it("reserves the approval transition before claiming or executing", async () => {
    const input = { teamId: "team-1", title: "Wait for persistence" };
    const messages = [approvalMessage(approvalPart({ input }))];

    const body = await createResponse({ messages })!.text();

    expect(mockBeginChatWrite).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "chat-1",
        messages,
        operation: "approval",
      }),
    );
    expect(mockBeginChatWrite.mock.invocationCallOrder[0]).toBeLessThan(
      mockClaimMutationApproval.mock.invocationCallOrder[0],
    );
    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(body).toContain("tool-output-available");
  });

  it("does not start or quarantine a mutation canceled after claim", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    const abortController = new AbortController();
    mockClaimMutationApproval.mockImplementationOnce(async () => {
      abortController.abort(new Error("client disconnected"));
      return { leaseToken: LEASE_TOKEN, state: "claimed" };
    });

    await createResponse({
      abortSignal: abortController.signal,
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Canceled before start" },
          }),
        ),
      ],
    })!.text();

    expect(mockStartMutationApproval).not.toHaveBeenCalled();
    expect(mockFailMutationApproval).not.toHaveBeenCalled();
    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    consoleError.mockRestore();
  });

  it("retries the durable claim when approval beats initial chat persistence", async () => {
    const notFoundError = Object.assign(new Error("Chat not found"), {
      status: 404,
    });
    mockClaimMutationApproval
      .mockRejectedValueOnce(notFoundError)
      .mockResolvedValueOnce({ leaseToken: LEASE_TOKEN, state: "claimed" });
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Fast approval" },
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockClaimMutationApproval).toHaveBeenCalledTimes(2);
    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(body).toContain("tool-output-available");
  });

  it("fails closed when another instance already holds the durable claim", async () => {
    mockClaimMutationApproval.mockResolvedValue({ state: "in_progress" });
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Only one executor" },
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    expect(mockCompleteMutationApproval).not.toHaveBeenCalled();
    expect(body).toContain("still being processed");
  });

  it("reclaims an expired ready lease before executing", async () => {
    mockClaimMutationApproval
      .mockResolvedValueOnce({
        leaseExpiresAt: new Date(Date.now() - 1000).toISOString(),
        state: "ready",
      })
      .mockResolvedValueOnce({ leaseToken: LEASE_TOKEN, state: "claimed" });
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Reclaimed action" },
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockClaimMutationApproval).toHaveBeenCalledTimes(2);
    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(body).toContain("tool-output-available");
  });

  it("polls an executing approval and replays its durable completion", async () => {
    const output = { storyId: "story-1", success: true };
    mockClaimMutationApproval
      .mockResolvedValueOnce({
        leaseExpiresAt: new Date(Date.now() + 60_000).toISOString(),
        state: "executing",
      })
      .mockResolvedValueOnce({ output, state: "completed" });
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Wait for completion" },
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockClaimMutationApproval).toHaveBeenCalledTimes(2);
    expect(mockStartMutationApproval).not.toHaveBeenCalled();
    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    expect(body).toContain("story-1");
  });

  it("reclaims instead of poisoning a definite start conflict", async () => {
    const conflict = Object.assign(new Error("lease expired"), { status: 409 });
    mockClaimMutationApproval
      .mockResolvedValueOnce({ leaseToken: "lease-1", state: "claimed" })
      .mockResolvedValueOnce({ leaseToken: "lease-2", state: "claimed" });
    mockStartMutationApproval
      .mockRejectedValueOnce(conflict)
      .mockResolvedValueOnce({ state: "started" });
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Retry safe start" },
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockStartMutationApproval).toHaveBeenCalledTimes(2);
    expect(mockFailMutationApproval).not.toHaveBeenCalled();
    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(body).toContain("tool-output-available");
  });

  it("never executes when the durable start transition is ambiguous", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    mockStartMutationApproval.mockRejectedValue(
      new Error("start response was lost"),
    );
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Do not guess" },
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    expect(mockFailMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({
        failureCode: "start_transition_uncertain",
        leaseToken: LEASE_TOKEN,
      }),
    );
    expect(mockCompleteMutationApproval).not.toHaveBeenCalled();
    expect(body).toContain("could not verify whether");
    consoleError.mockRestore();
  });

  it("quarantines an exception after execution starts and never replays it", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    mockCreateStoryExecute.mockImplementation(async () => {
      throw new Error("connection dropped after request dispatch");
    });
    const messages = [
      approvalMessage(
        approvalPart({
          input: { teamId: "team-1", title: "Maybe created" },
        }),
      ),
    ];

    const firstBody = await createResponse({ messages })!.text();
    const replayBody = await createResponse({ messages })!.text();

    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(mockFailMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({
        failureCode: "completion_persistence_uncertain",
      }),
    );
    expect(mockCompleteMutationApproval).not.toHaveBeenCalled();
    expect(firstBody).toContain("blocked until this execution is reconciled");
    expect(replayBody).toContain("blocked until this execution is reconciled");
    consoleError.mockRestore();
  });

  it.each([
    {
      error: new Error("connection dropped after dispatch"),
      label: "generic transport",
    },
    {
      error: new ApiError("Upstream unavailable", 503, {
        data: null,
        error: { message: "Upstream unavailable" },
      }),
      label: "5xx API",
    },
  ])(
    "quarantines a swallowed $label failure instead of completing its ledger entry",
    async ({ error }) => {
      const consoleError = jest.spyOn(console, "error").mockImplementation();
      mockUpdateStoryExecute.mockImplementation(async () => {
        const result = getApiError(error);
        return {
          success: false,
          error: result.error?.message ?? "Mutation failed",
        };
      });

      const response = createResponse({
        messages: [
          approvalMessage(
            approvalPart({
              input: { storyId: "story-1", title: "Maybe updated" },
              toolName: "updateStory",
            }),
          ),
        ],
      });

      const body = await response!.text();

      expect(mockCompleteMutationApproval).not.toHaveBeenCalled();
      expect(mockFailMutationApproval).toHaveBeenCalledWith(
        expect.objectContaining({
          failureCode: "completion_persistence_uncertain",
        }),
      );
      expect(body).toContain("blocked until this execution is reconciled");
      consoleError.mockRestore();
    },
  );

  it("completes a swallowed definite 4xx business failure as an error output", async () => {
    const apiResponse = {
      data: null,
      error: { message: "The story title is invalid" },
    };
    mockUpdateStoryExecute.mockImplementation(async () => {
      const result = getApiError(
        new ApiError("Invalid story", 422, apiResponse),
      );
      return {
        success: false,
        error: result.error?.message ?? "Mutation failed",
      };
    });

    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { storyId: "story-1", title: "Invalid title" },
            toolName: "updateStory",
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockCompleteMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({
        output: {
          success: false,
          error: "The story title is invalid",
        },
      }),
    );
    expect(mockFailMutationApproval).not.toHaveBeenCalled();
    expect(body).toContain("The story title is invalid");
  });

  it("isolates uncertain API reports between concurrent approved executions", async () => {
    let markSuccessfulExecutionStarted: (() => void) | undefined;
    let releaseSuccessfulExecution: (() => void) | undefined;
    const successfulExecutionStarted = new Promise<void>((resolve) => {
      markSuccessfulExecutionStarted = resolve;
    });
    const successfulExecutionCanFinish = new Promise<void>((resolve) => {
      releaseSuccessfulExecution = resolve;
    });
    mockUpdateStoryExecute.mockImplementation(
      async (input: { title: string }) => {
        if (input.title === "Successful update") {
          markSuccessfulExecutionStarted?.();
          await successfulExecutionCanFinish;
          return { success: true };
        }

        const result = getApiError(
          new Error("connection dropped after dispatch"),
        );
        releaseSuccessfulExecution?.();
        return {
          success: false,
          error: result.error?.message ?? "Mutation failed",
        };
      },
    );
    const successfulBody = createResponse({
      chatId: "chat-successful",
      messages: [
        approvalMessage(
          approvalPart({
            input: { storyId: "story-1", title: "Successful update" },
            toolCallId: "call-successful",
            toolName: "updateStory",
          }),
        ),
      ],
    })!.text();

    await successfulExecutionStarted;

    const consoleError = jest.spyOn(console, "error").mockImplementation();
    const uncertainBody = createResponse({
      chatId: "chat-uncertain",
      messages: [
        approvalMessage(
          approvalPart({
            input: { storyId: "story-2", title: "Uncertain update" },
            toolCallId: "call-uncertain",
            toolName: "updateStory",
          }),
        ),
      ],
    })!.text();

    try {
      const [successfulResponseBody, uncertainResponseBody] = await Promise.all(
        [successfulBody, uncertainBody],
      );

      expect(mockCompleteMutationApproval).toHaveBeenCalledWith(
        expect.objectContaining({ toolCallId: "call-successful" }),
      );
      expect(mockCompleteMutationApproval).not.toHaveBeenCalledWith(
        expect.objectContaining({ toolCallId: "call-uncertain" }),
      );
      expect(mockFailMutationApproval).toHaveBeenCalledWith(
        expect.objectContaining({
          failureCode: "completion_persistence_uncertain",
          toolCallId: "call-uncertain",
        }),
      );
      expect(successfulResponseBody).toContain("tool-output-available");
      expect(uncertainResponseBody).toContain(
        "blocked until this execution is reconciled",
      );
    } finally {
      consoleError.mockRestore();
    }
  });

  it("bounds tool execution below the durable lease", async () => {
    jest.useFakeTimers();
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    let markExecutionStarted: (() => void) | undefined;
    const executionStarted = new Promise<void>((resolve) => {
      markExecutionStarted = resolve;
    });
    mockCreateStoryExecute.mockImplementation(
      async () =>
        new Promise(() => {
          markExecutionStarted?.();
        }),
    );

    try {
      const bodyPromise = createResponse({
        messages: [
          approvalMessage(
            approvalPart({
              input: { teamId: "team-1", title: "Bounded execution" },
            }),
          ),
        ],
      })!.text();
      await executionStarted;
      await jest.advanceTimersByTimeAsync(110_000);
      const body = await bodyPromise;

      expect(mockFailMutationApproval).toHaveBeenCalledWith(
        expect.objectContaining({
          failureCode: "completion_persistence_uncertain",
        }),
      );
      expect(body).toContain("blocked until this execution is reconciled");
    } finally {
      consoleError.mockRestore();
      jest.useRealTimers();
    }
  });

  it("quarantines a lost completion response instead of executing again", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    mockCompleteMutationApproval.mockRejectedValue(
      new Error("completion response was lost"),
    );
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "One dispatch only" },
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(mockFailMutationApproval).toHaveBeenCalledWith(
      expect.objectContaining({
        failureCode: "completion_persistence_uncertain",
      }),
    );
    expect(body).toContain("blocked until this execution is reconciled");
    consoleError.mockRestore();
  });

  it("re-reads durable completion when completion and failure responses are both lost", async () => {
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    const durableOutput = { storyId: "story-1", success: true };
    mockClaimMutationApproval
      .mockResolvedValueOnce({ leaseToken: LEASE_TOKEN, state: "claimed" })
      .mockResolvedValueOnce({
        output: durableOutput,
        state: "completed",
      });
    mockCompleteMutationApproval.mockRejectedValue(
      new Error("completion response was lost after commit"),
    );
    mockFailMutationApproval.mockRejectedValue(
      new Error("failure response was also lost"),
    );

    const body = await createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "One durable dispatch" },
          }),
        ),
      ],
    })!.text();

    expect(mockCreateStoryExecute).toHaveBeenCalledTimes(1);
    expect(mockClaimMutationApproval).toHaveBeenCalledTimes(2);
    expect(body).toContain("story-1");
    expect(body).not.toContain("blocked until this execution is reconciled");
    consoleError.mockRestore();
  });

  it("returns a terminal uncertain result without starting the tool", async () => {
    mockClaimMutationApproval.mockResolvedValue({
      failureCode: "execution_lease_expired",
      state: "failed_uncertain",
    });
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Needs reconciliation" },
          }),
        ),
      ],
    });

    const body = await response!.text();

    expect(mockStartMutationApproval).not.toHaveBeenCalled();
    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    expect(body).toContain("blocked until this execution is reconciled");
  });

  it("returns 401 without executing when no authenticated user is present", async () => {
    const response = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: { teamId: "team-1", title: "Unauthorized" },
          }),
        ),
      ],
      userId: "",
    });

    expect(response?.status).toBe(401);
    await expect(response!.text()).resolves.toBe("Unauthorized");
    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
  });

  it("does not intercept setup tools or ordinary chat messages", () => {
    const setupResponse = createResponse({
      messages: [
        approvalMessage(
          approvalPart({
            input: {},
            toolName: "createGitHubInstallSessionTool",
          }),
        ),
      ],
    });
    const ordinaryResponse = createResponse({
      messages: [
        {
          id: "user-1",
          parts: [{ text: "Create a story", type: "text" }],
          role: "user",
        },
      ],
    });

    expect(setupResponse).toBeUndefined();
    expect(ordinaryResponse).toBeUndefined();
    expect(mockGitHubSetupExecute).not.toHaveBeenCalled();
  });
});
