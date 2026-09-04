/* global beforeAll, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import {
  ReadableStream as NodeReadableStream,
  TransformStream as NodeTransformStream,
  WritableStream as NodeWritableStream,
} from "node:stream/web";
import type { NextRequest } from "next/server";
import type * as ChatRequestModule from "./chat-request";

jest.mock("server-only", () => ({}));

let chatRequestModule: typeof ChatRequestModule;

const createRequest = (body: unknown) =>
  ({
    headers: new Headers(),
    text: async () => JSON.stringify(body),
  }) satisfies Pick<NextRequest, "headers" | "text">;

const textMessage = {
  id: "user-1",
  parts: [{ text: "Summarize this workspace", type: "text" }],
  role: "user",
};

describe("Maya chat request decoding", () => {
  beforeAll(async () => {
    globalThis.ReadableStream =
      NodeReadableStream as typeof globalThis.ReadableStream;
    globalThis.TransformStream =
      NodeTransformStream as typeof globalThis.TransformStream;
    globalThis.WritableStream =
      NodeWritableStream as typeof globalThis.WritableStream;
    const { Response: EdgeResponse } = await import(
      "next/dist/compiled/@edge-runtime/primitives/fetch"
    );
    globalThis.Response = EdgeResponse;
    chatRequestModule = await import("./chat-request");
  });

  it.each([
    ["a null message list", null],
    [
      "null assistant parts",
      [{ id: "assistant-1", parts: null, role: "assistant" }],
    ],
    ["a null part", [{ id: "assistant-1", parts: [null], role: "assistant" }]],
    [
      "a part with a null type",
      [{ id: "assistant-1", parts: [{ type: null }], role: "assistant" }],
    ],
    [
      "a null approval",
      [
        {
          id: "assistant-1",
          parts: [
            {
              approval: null,
              input: { teamId: "team-1", title: "Unsafe" },
              state: "approval-responded",
              toolCallId: "call-1",
              type: "tool-createStory",
            },
          ],
          role: "assistant",
        },
      ],
    ],
  ])(
    "returns 400 for %s without dispatching downstream work",
    async (_label, messages) => {
      const handle = jest.fn(async () => new Response("stream"));

      const response = await chatRequestModule.dispatchValidatedChatRequest({
        handle,
        request: createRequest({ messages }),
      });

      expect(response.status).toBe(400);
      await expect(response.text()).resolves.toBe("Invalid Maya chat request.");
      expect(handle).not.toHaveBeenCalled();
    },
  );

  it("preserves structurally valid historical, custom-data, and approval parts", async () => {
    const messages = [
      textMessage,
      {
        id: "assistant-1",
        parts: [
          {
            input: { legacy: true },
            output: { success: true },
            state: "output-available",
            toolCallId: "legacy-call",
            type: "tool-retiredTool",
          },
          {
            data: { value: 1 },
            id: "custom-data-1",
            type: "data-custom",
          },
          {
            approval: { approved: true, id: "approval-1" },
            input: { teamId: "team-1", title: "Validated later" },
            state: "approval-responded",
            toolCallId: "call-1",
            type: "tool-createStory",
          },
        ],
        role: "assistant",
      },
    ];

    const handle = jest.fn(
      async (_requestBody: ChatRequestModule.ChatRequestBody) =>
        new Response("stream"),
    );
    const response = await chatRequestModule.dispatchValidatedChatRequest({
      handle,
      request: createRequest({ messages }),
    });

    expect(response.status).toBe(200);
    expect(handle).toHaveBeenCalledTimes(1);
    expect(handle.mock.calls[0]?.[0].messages).toEqual(messages);
  });

  it("accepts a bounded Google Drive reference selected on the user turn", async () => {
    const messages = [
      {
        id: "user-drive-context",
        parts: [
          { text: "Review this file", type: "text" },
          {
            data: {
              referenceId: "00000000-0000-4000-8000-000000000001",
              name: "Launch plan",
              mimeType: "application/vnd.google-apps.document",
            },
            type: "data-google-drive-file-context",
          },
        ],
        role: "user",
      },
    ];
    const handle = jest.fn(async () => new Response("stream"));

    const response = await chatRequestModule.dispatchValidatedChatRequest({
      handle,
      request: createRequest({ messages }),
    });

    expect(response.status).toBe(200);
    expect(handle).toHaveBeenCalledTimes(1);
  });

  it("rejects a Google Drive provider ID in place of an opaque reference", async () => {
    const messages = [
      {
        id: "user-drive-context",
        parts: [
          { text: "Review this file", type: "text" },
          {
            data: {
              referenceId: "google-provider-file-id",
              name: "Launch plan",
              mimeType: "application/vnd.google-apps.document",
            },
            type: "data-google-drive-file-context",
          },
        ],
        role: "user",
      },
    ];
    const handle = jest.fn(async () => new Response("stream"));

    const response = await chatRequestModule.dispatchValidatedChatRequest({
      handle,
      request: createRequest({ messages }),
    });

    expect(response.status).toBe(400);
    expect(handle).not.toHaveBeenCalled();
  });

  it("rejects a selected Drive file Maya cannot read as text", async () => {
    const messages = [
      {
        id: "user-drive-context",
        parts: [
          { text: "Review this file", type: "text" },
          {
            data: {
              referenceId: "00000000-0000-4000-8000-000000000001",
              name: "Brand image",
              mimeType: "image/png",
            },
            type: "data-google-drive-file-context",
          },
        ],
        role: "user",
      },
    ];
    const handle = jest.fn(async () => new Response("stream"));

    const response = await chatRequestModule.dispatchValidatedChatRequest({
      handle,
      request: createRequest({ messages }),
    });

    expect(response.status).toBe(400);
    expect(handle).not.toHaveBeenCalled();
  });
});
