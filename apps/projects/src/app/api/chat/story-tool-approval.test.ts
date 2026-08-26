/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

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
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { createStory } from "@/lib/ai/tools/stories/create-story";
import { bulkCreateStories } from "@/lib/ai/tools/stories/bulk-create-stories";
import { saveChat } from "./save-chat";

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

const loadStoryToolApproval = async () => {
  const { Response: EdgeResponse } = await import(
    "next/dist/compiled/@edge-runtime/primitives/fetch"
  );
  globalThis.Response = EdgeResponse;

  return import("./story-tool-approval");
};

jest.mock("@/lib/ai/tools/stories/create-story", () => ({
  createStory: {
    execute: jest.fn(async (input: { title: string }) => ({
      message: `Story "${input.title}" created successfully.`,
      success: true,
    })),
  },
  createStoryInputSchema: { parse: (input: unknown) => input },
}));

jest.mock("@/lib/ai/tools/stories/bulk-create-stories", () => ({
  bulkCreateStories: { execute: jest.fn() },
  bulkCreateStoriesInputSchema: { parse: (input: unknown) => input },
}));

jest.mock("./save-chat", () => ({ saveChat: jest.fn() }));

const mockCreateStoryExecute = jest.mocked(createStory.execute!);
const mockBulkCreateStoriesExecute = jest.mocked(bulkCreateStories.execute!);
const mockSaveChat = jest.mocked(saveChat);

const approvalMessage = ({
  approved,
  input,
  type = "tool-createStory",
}: {
  approved: boolean;
  input: unknown;
  type?: "tool-bulkCreateStories" | "tool-createStory";
}) =>
  ({
    id: "assistant-1",
    parts: [
      {
        approval: { approved, id: "approval-1" },
        input,
        state: "approval-responded",
        toolCallId: "call-1",
        type,
      },
    ],
    role: "assistant",
  }) as unknown as MayaUIMessage;

describe("story tool approval response", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("executes the exact approved story payload without another model call", async () => {
    const { createStoryToolApprovalResponse } = await loadStoryToolApproval();
    const messages = [
      approvalMessage({
        approved: true,
        input: { teamId: "team-1", title: "Add onboarding checklist" },
      }),
    ];

    const response = createStoryToolApprovalResponse({
      abortSignal: new AbortController().signal,
      chatId: "chat-1",
      messages,
      workspaceSlug: "acme",
    });

    expect(response).toBeDefined();
    const streamBody = await response!.text();

    expect(mockCreateStoryExecute).toHaveBeenCalledWith(
      { teamId: "team-1", title: "Add onboarding checklist" },
      expect.objectContaining({
        experimental_context: { workspaceSlug: "acme" },
        toolCallId: "call-1",
      }),
    );
    expect(mockBulkCreateStoriesExecute).not.toHaveBeenCalled();
    expect(streamBody).toContain("tool-output-available");
    expect(streamBody).toContain("created successfully");
    expect(mockSaveChat).toHaveBeenCalledWith(
      expect.objectContaining({
        generateTitle: false,
        id: "chat-1",
        workspaceSlug: "acme",
      }),
    );
  });

  it("records a denial without executing the story tool", async () => {
    const { createStoryToolApprovalResponse } = await loadStoryToolApproval();
    const response = createStoryToolApprovalResponse({
      abortSignal: new AbortController().signal,
      chatId: "chat-1",
      messages: [
        approvalMessage({
          approved: false,
          input: { teamId: "team-1", title: "Do not create this" },
        }),
      ],
      workspaceSlug: "acme",
    });

    const streamBody = await response!.text();

    expect(mockCreateStoryExecute).not.toHaveBeenCalled();
    expect(streamBody).toContain("tool-output-denied");
    expect(mockSaveChat).toHaveBeenCalled();
  });

  it("ignores ordinary chat messages", async () => {
    const { createStoryToolApprovalResponse } = await loadStoryToolApproval();
    const response = createStoryToolApprovalResponse({
      abortSignal: new AbortController().signal,
      chatId: "chat-1",
      messages: [
        {
          id: "user-1",
          parts: [{ text: "Create a story", type: "text" }],
          role: "user",
        },
      ],
      workspaceSlug: "acme",
    });

    expect(response).toBeUndefined();
  });
});
